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

	for _, file := range files {
		fileType := "文件"
		if file.IsDir() {
			fileType = "目录"
		}

		result.WriteString(fmt.Sprintf("%s\t%s\t%d bytes\n", file.Name(), fileType, file.Size()))
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