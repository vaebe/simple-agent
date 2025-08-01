package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ExecuteFileOperation 执行文件操作
func ExecuteFileOperation(tool Tool, mcpClient interface{}) ToolCallResponse {
	switch tool.Name {
	case "list":
		// 获取目录参数
		dir, ok := tool.Args["path"].(string)
		if !ok {
			dir = "."
		}

		// 安全检查：限制目录访问范围
		if strings.Contains(dir, "..") || strings.HasPrefix(dir, "/") {
			return ToolCallResponse{Error: "出于安全考虑，禁止访问上级目录或绝对路径"}
		}

		// 列出目录内容
		result, err := listDirectory(dir)
		if err != nil {
			return ToolCallResponse{Error: err.Error()}
		}
		return ToolCallResponse{Result: result}

	case "read":
		// 获取文件路径参数
		path, ok := tool.Args["path"].(string)
		if !ok {
			return ToolCallResponse{Error: "缺少文件路径参数"}
		}

		// 安全检查：限制文件访问范围
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			return ToolCallResponse{Error: "出于安全考虑，禁止访问上级目录或绝对路径"}
		}

		// 读取文件内容
		result, err := readFile(path)
		if err != nil {
			return ToolCallResponse{Error: err.Error()}
		}
		return ToolCallResponse{Result: result}

	case "write":
		// 获取文件路径参数
		path, ok := tool.Args["path"].(string)
		if !ok {
			return ToolCallResponse{Error: "缺少文件路径参数"}
		}

		// 获取文件内容参数
		content, ok := tool.Args["content"].(string)
		if !ok {
			return ToolCallResponse{Error: "缺少文件内容参数"}
		}

		// 安全检查：限制文件写入范围
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			return ToolCallResponse{Error: "出于安全考虑，禁止写入上级目录或绝对路径"}
		}

		// 写入文件内容
		err := writeFile(path, content)
		if err != nil {
			return ToolCallResponse{Error: err.Error()}
		}
		return ToolCallResponse{Result: fmt.Sprintf("成功写入文件 %s，内容长度: %d", path, len(content))}

	default:
		return ToolCallResponse{Error: fmt.Sprintf("未知的文件操作: %s", tool.Name)}
	}
}

// listDirectory 列出目录内容
func listDirectory(path string) (string, error) {
	// 确保路径存在
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("无法访问路径 %s: %v", path, err)
	}

	// 检查是否是目录
	if !info.IsDir() {
		return "", fmt.Errorf("%s 不是一个目录", path)
	}

	// 读取目录内容
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("无法读取目录 %s: %v", path, err)
	}

	// 格式化输出
	var result strings.Builder
	result.WriteString(fmt.Sprintf("目录 %s 的内容:\n", path))
	result.WriteString("名称\t类型\t大小\n")
	result.WriteString("----\t----\t----\n")

	for _, file := range files {
		fileType := "文件"
		if file.IsDir() {
			fileType = "目录"
		}

		size := file.Size()
		sizeStr := fmt.Sprintf("%d bytes", size)
		if size > 1024*1024 {
			sizeStr = fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
		} else if size > 1024 {
			sizeStr = fmt.Sprintf("%.2f KB", float64(size)/1024)
		}

		result.WriteString(fmt.Sprintf("%s\t%s\t%s\n", file.Name(), fileType, sizeStr))
	}

	return result.String(), nil
}

// readFile 读取文件内容
func readFile(path string) (string, error) {
	// 确保路径存在
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("无法访问文件 %s: %v", path, err)
	}

	// 检查是否是文件
	if info.IsDir() {
		return "", fmt.Errorf("%s 是一个目录，不是文件", path)
	}

	// 检查文件大小（限制为1MB）
	const maxFileSize = 1024 * 1024 // 1MB
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("文件 %s 太大 (%d bytes)，最大支持 1MB", path, info.Size())
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("无法读取文件 %s: %v", path, err)
	}

	return string(content), nil
}

// writeFile 写入文件内容
func writeFile(path string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目录 %s: %v", dir, err)
	}

	// 写入文件内容
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("无法写入文件 %s: %v", path, err)
	}

	return nil
}
