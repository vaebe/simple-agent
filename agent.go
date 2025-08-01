package main

import (
	"context"
	"fmt"
	"simple-agent/logger"
	"simple-agent/tools"
	"strings"

	"github.com/imroc/req/v3"
	"go.uber.org/zap"
)

// Message 表示与模型交互的消息结构
type Message struct {
	Role    string `json:"role"`    // 消息角色：system, user, assistant, tool
	Content string `json:"content"` // 消息内容
}

const GLM_API_URL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

// 预定义的系统提示词
const DEFAULT_SYSTEM_PROMPT = `你是一个智能助手，由智谱AI的GLM-4.5-Flash模型提供支持。
你拥有强大的推理能力、稳定的代码生成和多工具协同处理能力，同时具备显著的运行速度优势。
你支持最长128K的上下文处理，可高效应对长文本理解、多轮对话连续性和结构化内容生成等复杂任务。

你可以回答用户的问题，也可以使用工具来完成特定任务。

重要：当你需要使用工具时，必须严格按照以下JSON格式返回，不要添加任何其他内容：

[{
  "type": "工具类型",
  "name": "工具名称", 
  "args": {
    "参数名": "参数值"
  },
  "thought": "思考过程"
}]

可用的工具：

1. 文件操作工具 (file_operation)：
   - list: 列出目录内容，参数：{"path": "目录路径"}
   - read: 读取文件内容，参数：{"path": "文件路径"}
   - write: 写入文件内容，参数：{"path": "文件路径", "content": "文件内容"}

2. Shell命令工具 (shell_command)：
   - execute: 执行Shell命令，参数：{"command": "要执行的命令"}

使用工具的规则：
- 当前工作目录是项目的根目录，你可以直接使用相对路径访问项目文件
- 当用户要求分析代码时，首先使用list工具查看项目结构，然后使用read工具读取相关文件
- 当用户要求创建、读取、修改文件时，使用文件操作工具
- 当用户要求执行命令或运行程序时，使用Shell命令工具
- 总是先思考为什么需要使用工具，然后在thought字段中说明
- 工具调用必须使用正确的JSON格式，不要添加任何解释文字
- 当工具执行完成后，如果用户明确要求基于工具结果提供分析，你必须直接提供分析结果，不能再返回工具调用格式
- 如果不需要使用工具，直接回答用户问题

示例：
当用户说"分析这个仓库的代码有什么可以优化的地方"时，你应该先使用list工具查看项目结构：

[{
  "type": "file_operation",
  "name": "list",
  "args": {
    "path": "."
  },
  "thought": "用户要求分析代码优化点，我需要先查看项目的文件结构，了解有哪些文件可以分析"
}]

当工具执行完成后，如果用户说"基于上述工具执行结果，请为我提供详细的代码分析报告"，你必须直接提供分析结果，不能再返回工具调用格式。

注意：工具调用时只返回JSON格式，不要添加任何其他文字说明。如果用户输入为空，请友好地提示用户输入内容。`

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
		// 获取用户输入（readline已经处理了提示符）
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

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

		// 处理多轮工具调用
		for {
			// 检查回复中是否包含工具调用
			extractedTools, hasTools := tools.ExtractTools(message.Content)
			if !hasTools {
				break // 没有工具调用，退出循环
			}

			logger.Info("检测到工具调用", zap.Int("数量", len(extractedTools)))

			// 处理工具调用
			responses := tools.ExecuteTools(ctx, extractedTools, a)

			logger.Info("工具调用执行完成", zap.Int("响应数量", len(responses)))

			// 将工具调用结果添加到对话历史
			toolResponseMsg := Message{
				Role:    "tool",
				Content: tools.FormatToolResponses(responses),
			}
			a.conversation = append(a.conversation, toolResponseMsg)

			logger.Info("工具响应已添加到对话历史")
			logger.Debug("工具响应内容", zap.String("内容", toolResponseMsg.Content))

			// 再次调用模型获取回复
			message, err = a.runInference(ctx, a.conversation)
			if err != nil {
				logger.Error("模型推理失败", zap.Error(err))
				fmt.Printf("\u001b[91m错误\u001b[0m: %v\n", err)
				break
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
	logger.Info("正在发送请求到智谱API")
	fmt.Println("正在发送请求到智谱API...")
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.config.APIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(GLM_API_URL)

	if err != nil {
		logger.Error("API请求错误", zap.Error(err))
		fmt.Printf("API请求错误: %v\n", err)
		return Message{}, err
	}

	// 打印响应状态码
	logger.Info("API响应状态码", zap.Int("状态码", resp.StatusCode))
	fmt.Printf("API响应状态码: %d\n", resp.StatusCode)

	// 获取原始响应内容用于调试
	rawBody, _ := resp.ToBytes()

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		logger.Error("API返回错误状态码", zap.Int("状态码", resp.StatusCode), zap.String("错误响应", string(rawBody)))
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
		logger.Error("解析API响应失败", zap.Error(err), zap.String("原始响应", string(rawBody)))
		fmt.Printf("解析API响应失败: %v\n", err)
		return Message{}, fmt.Errorf("解析API响应失败: %v, 原始响应: %s", err, string(rawBody))
	}

	// 检查是否有API错误
	if response.Error.Message != "" {
		logger.Error("API错误", zap.String("消息", response.Error.Message), zap.String("类型", response.Error.Type), zap.String("代码", response.Error.Code))
		fmt.Printf("API错误: %s (类型: %s, 代码: %s)\n",
			response.Error.Message, response.Error.Type, response.Error.Code)
		return Message{}, fmt.Errorf("API错误: %s (类型: %s, 代码: %s)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// 检查是否有有效回复
	if len(response.Choices) == 0 {
		logger.Error("API返回了空的choices数组")
		fmt.Println("API返回了空的choices数组")
		return Message{}, fmt.Errorf("API返回了空回复")
	}

	// 打印模型返回的原始内容用于调试
	logger.Debug("模型返回的原始内容", zap.String("内容", response.Choices[0].Message.Content))
	fmt.Printf("\u001b[96m调试\u001b[0m: 模型返回的原始内容:\n%s\n", response.Choices[0].Message.Content)

	return response.Choices[0].Message, nil
}
