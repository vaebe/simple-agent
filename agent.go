package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/imroc/req/v3"
)

// 定义工具类型常量
const (
	TOOL_FILE_OPERATION = "file_operation" // 文件操作工具
	TOOL_SHELL_COMMAND  = "shell_command"  // Shell命令工具
)

// AgentConfig 代理配置
type AgentConfig struct {
	APIKey       string   // 智谱API密钥
	SystemPrompt string   // 系统提示词
	Tools        []string // 可用工具列表
}

// AdvancedAgent 高级代理结构体
type AdvancedAgent struct {
	config         AgentConfig           // 代理配置
	getUserMessage func() (string, bool) // 获取用户消息的函数
	mcpClient      *MCPClient            // 文件操作客户端
	conversation   []Message             // 对话历史
}

// NewAdvancedAgent 创建一个新的高级代理实例
func NewAdvancedAgent(config AgentConfig, getUserMessage func() (string, bool)) *AdvancedAgent {
	// 初始化对话历史，添加系统提示
	conversation := []Message{
		{Role: "system", Content: config.SystemPrompt},
	}

	return &AdvancedAgent{
		config:         config,
		getUserMessage: getUserMessage,
		mcpClient:      NewMCPClient(),
		conversation:   conversation,
	}
}

// Tool 工具结构体
type Tool struct {
	Type    string                 `json:"type"`    // 工具类型
	Name    string                 `json:"name"`    // 工具名称
	Args    map[string]interface{} `json:"args"`    // 工具参数
	Thought string                 `json:"thought"` // 工具思考过程
}

// ToolCallResponse 工具调用响应
type ToolCallResponse struct {
	Result string `json:"result"` // 工具调用结果
	Error  string `json:"error"`  // 错误信息（如果有）
}

// Run 运行高级代理的主循环
func (a *AdvancedAgent) Run(ctx context.Context) error {
	fmt.Println("与GLM-4.5聊天 (使用'ctrl-c'退出)")
	fmt.Println("可用工具: " + strings.Join(a.config.Tools, ", "))

	for {
		// 提示用户输入
		fmt.Print("\u001b[94m你\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		// 检查是否是工具调用
		if strings.HasPrefix(userInput, "/tool") {
			// 处理工具调用
			a.handleToolCall(ctx, userInput)
			continue
		}

		// 添加用户消息到对话历史
		userMessage := Message{Role: "user", Content: userInput}
		a.conversation = append(a.conversation, userMessage)

		// 调用模型获取回复
		message, err := a.runInference(ctx, a.conversation)
		if err != nil {
			return err
		}

		// 检查回复中是否包含工具调用
		tools, hasTools := a.extractTools(message.Content)
		if hasTools {
			// 处理工具调用
			responses := a.executeTools(ctx, tools)

			// 将工具调用结果添加到对话历史
			toolResponseMsg := Message{
				Role:    "tool",
				Content: a.formatToolResponses(responses),
			}
			a.conversation = append(a.conversation, toolResponseMsg)

			// 再次调用模型获取最终回复
			message, err = a.runInference(ctx, a.conversation)
			if err != nil {
				return err
			}
		}

		// 添加模型回复到对话历史
		a.conversation = append(a.conversation, message)

		// 打印模型回复
		fmt.Printf("\u001b[93mGLM-4.5-Flash\u001b[0m: %s\n", message.Content)
	}

	return nil
}

// runInference 调用智谱API获取模型回复
func (a *AdvancedAgent) runInference(ctx context.Context, conversation []Message) (Message, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    "glm-4.5", // 使用GLM-4.5模型
		"messages": conversation,
		"thinking": map[string]string{
			"type": "enabled", // 启用动态思考模式
		},
	}

	// 创建HTTP客户端
	client := req.C()

	// 发送请求到智谱API
	fmt.Println("正在发送请求到智谱API...")
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.config.APIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(GLM_API_URL)

	if err != nil {
		fmt.Printf("API请求错误: %v\n", err)
		return Message{}, err
	}

	// 打印响应状态码
	fmt.Printf("API响应状态码: %d\n", resp.StatusCode)

	// 获取原始响应内容用于调试
	rawBody, _ := resp.ToBytes()
	// fmt.Println("API响应内容:")
	fmt.Println(string(rawBody))

	// 解析响应
	var response struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	err = resp.UnmarshalJson(&response)
	if err != nil {
		fmt.Printf("解析API响应失败: %v\n", err)
		return Message{}, fmt.Errorf("解析API响应失败: %v, 原始响应: %s", err, string(rawBody))
	}

	// 检查是否有API错误
	if response.Error.Message != "" {
		fmt.Printf("API错误: %s (类型: %s, 代码: %s)\n",
			response.Error.Message, response.Error.Type, response.Error.Code)
		return Message{}, fmt.Errorf("API错误: %s (类型: %s, 代码: %s)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// 检查是否有有效回复
	if len(response.Choices) == 0 {
		fmt.Println("API返回了空的choices数组")
		return Message{}, fmt.Errorf("API返回了空回复")
	}

	return response.Choices[0].Message, nil
}

// handleToolCall 处理用户直接输入的工具调用命令
func (a *AdvancedAgent) handleToolCall(ctx context.Context, command string) {
	parts := strings.Fields(command)
	if len(parts) < 3 {
		fmt.Println("错误: 无效的工具命令，格式应为 /tool <工具类型> <工具名称> [参数]")
		return
	}

	toolType := parts[1]
	toolName := parts[2]

	// 解析参数
	args := make(map[string]interface{})
	if len(parts) > 3 {
		argStr := strings.Join(parts[3:], " ")

		err := json.Unmarshal([]byte(argStr), &args)
		if err != nil {
			fmt.Printf("错误: 无法解析参数: %s\n", err)
			return
		}
	}

	// 创建工具调用
	tool := Tool{
		Type: toolType,
		Name: toolName,
		Args: args,
	}

	// 执行工具调用
	response := a.executeTool(ctx, tool)

	// 打印结果
	fmt.Printf("\u001b[93m工具结果\u001b[0m: %s\n", response.Result)
	if response.Error != "" {
		fmt.Printf("\u001b[91m错误\u001b[0m: %s\n", response.Error)
	}
}

// extractTools 从模型回复中提取工具调用
func (a *AdvancedAgent) extractTools(content string) ([]Tool, bool) {
	// 查找工具调用的JSON格式
	start := strings.Index(content, "```json")
	if start == -1 {
		return nil, false
	}

	end := strings.Index(content[start:], "```")
	if end == -1 {
		return nil, false
	}

	// 提取JSON内容
	jsonStr := content[start+7 : start+end]

	// 解析JSON
	var tools []Tool
	err := json.Unmarshal([]byte(jsonStr), &tools)
	if err != nil {
		fmt.Printf("错误: 无法解析工具调用: %s\n", err)
		return nil, false
	}

	return tools, len(tools) > 0
}

// executeTools 执行多个工具调用
func (a *AdvancedAgent) executeTools(ctx context.Context, tools []Tool) []ToolCallResponse {
	responses := make([]ToolCallResponse, len(tools))

	for i, tool := range tools {
		responses[i] = a.executeTool(ctx, tool)
	}

	return responses
}

// executeTool 执行单个工具调用
func (a *AdvancedAgent) executeTool(ctx context.Context, tool Tool) ToolCallResponse {
	switch tool.Type {
	case TOOL_FILE_OPERATION:
		return a.executeFileOperation(tool)

	case TOOL_SHELL_COMMAND:
		return a.executeShellCommand(tool)

	default:
		return ToolCallResponse{
			Error: fmt.Sprintf("未知的工具类型: %s", tool.Type),
		}
	}
}

// executeFileOperation 执行文件操作
func (a *AdvancedAgent) executeFileOperation(tool Tool) ToolCallResponse {
	switch tool.Name {
	case "list":
		// 获取目录参数
		dir, ok := tool.Args["path"].(string)
		if !ok {
			dir = "."
		}

		// 列出目录内容
		result, err := a.mcpClient.ListDirectory(dir)
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

		// 读取文件内容
		result, err := a.mcpClient.ReadFile(path)
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

		// 写入文件内容
		err := a.mcpClient.WriteFile(path, content)
		if err != nil {
			return ToolCallResponse{Error: err.Error()}
		}
		return ToolCallResponse{Result: fmt.Sprintf("成功写入文件 %s", path)}

	default:
		return ToolCallResponse{Error: fmt.Sprintf("未知的文件操作: %s", tool.Name)}
	}
}

// executeShellCommand 执行Shell命令
func (a *AdvancedAgent) executeShellCommand(tool Tool) ToolCallResponse {
	// 获取命令参数
	cmdStr, ok := tool.Args["command"].(string)
	if !ok {
		return ToolCallResponse{Error: "缺少命令参数"}
	}

	// 创建命令
	cmd := exec.Command("sh", "-c", cmdStr)

	// 获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolCallResponse{
			Result: string(output),
			Error:  err.Error(),
		}
	}

	return ToolCallResponse{Result: string(output)}
}

// formatToolResponses 格式化工具调用结果
func (a *AdvancedAgent) formatToolResponses(responses []ToolCallResponse) string {
	result := ""

	for i, resp := range responses {
		result += fmt.Sprintf("工具调用 %d 结果:\n", i+1)
		result += fmt.Sprintf("输出: %s\n", resp.Result)

		if resp.Error != "" {
			result += fmt.Sprintf("错误: %s\n", resp.Error)
		}

		result += "\n"
	}

	return result
}
