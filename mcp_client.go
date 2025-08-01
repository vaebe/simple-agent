package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// MCPClient 文件操作服务客户端
type MCPClient struct {
	// 可以添加配置选项，如API密钥、服务地址等
}

// NewMCPClient 创建一个新的MCP客户端
func NewMCPClient() *MCPClient {
	return &MCPClient{}
}

// ListDirectory 列出目录内容
func (c *MCPClient) ListDirectory(path string) (string, error) {
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

// ReadFile 读取文件内容
func (c *MCPClient) ReadFile(path string) (string, error) {
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

// WriteFile 写入文件内容
func (c *MCPClient) WriteFile(path string, content string) error {
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