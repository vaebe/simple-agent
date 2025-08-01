package main

import (
	"context"
	"fmt"
	"simple-agent/tools"
	"strings"

	"github.com/imroc/req/v3"
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
	conversation   []Message             // 对话历史
}

// 实现ToolExecutor接口
func (a *AdvancedAgent) ExecuteFileOperation(tool tools.Tool) tools.ToolCallResponse {
	// 直接调用tools包中的文件操作实现
	return tools.ExecuteFileOperation(tool)
}

func (a *AdvancedAgent) ExecuteShellCommand(tool tools.Tool) tools.ToolCallResponse {
	return tools.ExecuteShellCommand(tool)
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
		conversation:   conversation,
	}
}

// Run 运行高级代理的主循环
func (a *AdvancedAgent) Run(ctx context.Context) error {
	fmt.Println("与GLM-4.5-Flash聊天 (使用'ctrl-c'退出)")
	fmt.Println("可用工具: " + strings.Join(a.config.Tools, ", "))

	for {
		// 提示用户输入
		fmt.Print("\u001b[94m你\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		// 清除当前行，重新打印用户输入（用于更好的显示效果）
		fmt.Printf("\r\u001b[94m你\u001b[0m: %s\n", userInput)

		// 检查是否是工具调用
		if strings.HasPrefix(userInput, "/tool") {
			// 处理工具调用
			tools.HandleToolCall(ctx, userInput, a)
			continue
		}

		// 添加用户消息到对话历史
		userMessage := Message{Role: "user", Content: userInput}
		a.conversation = append(a.conversation, userMessage)

		// 调用模型获取回复
		message, err := a.runInference(ctx, a.conversation)
		if err != nil {
			fmt.Printf("\u001b[91m错误\u001b[0m: %v\n", err)
			continue
		}

		// 检查回复中是否包含工具调用
		extractedTools, hasTools := tools.ExtractTools(message.Content)
		if hasTools {
			// 处理工具调用
			responses := tools.ExecuteTools(ctx, extractedTools, a)

			// 将工具调用结果添加到对话历史
			toolResponseMsg := Message{
				Role:    "tool",
				Content: tools.FormatToolResponses(responses),
			}
			a.conversation = append(a.conversation, toolResponseMsg)

			// 再次调用模型获取最终回复
			message, err = a.runInference(ctx, a.conversation)
			if err != nil {
				fmt.Printf("\u001b[91m错误\u001b[0m: %v\n", err)
				continue
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
func (a *AdvancedAgent) runInference(_ctx context.Context, conversation []Message) (Message, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    "glm-4.5", // 使用GLM-4.5-Flash模型
		"messages": conversation,
		"thinking": map[string]string{
			"type": "enabled", // 启用动态思考模式
		},
		"stream": false, // 暂时不使用流式输出
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

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		fmt.Printf("API返回错误状态码: %d\n", resp.StatusCode)
		fmt.Printf("错误响应: %s\n", string(rawBody))
		return Message{}, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

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
