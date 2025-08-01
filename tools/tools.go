package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"simple-agent/logger"

	"go.uber.org/zap"
)

// ToolExecutor 工具执行器接口
type ToolExecutor interface {
	ExecuteFileOperation(tool Tool) ToolCallResponse
	ExecuteShellCommand(tool Tool) ToolCallResponse
}

// ExecuteTool 执行单个工具调用
func ExecuteTool(_ctx context.Context, tool Tool, executor ToolExecutor) ToolCallResponse {
	switch tool.Type {
	case TOOL_FILE_OPERATION:
		return executor.ExecuteFileOperation(tool)

	case TOOL_SHELL_COMMAND:
		return executor.ExecuteShellCommand(tool)

	default:
		return ToolCallResponse{
			Error: fmt.Sprintf("未知的工具类型: %s", tool.Type),
		}
	}
}

// ExecuteTools 执行多个工具调用
func ExecuteTools(ctx context.Context, tools []Tool, executor ToolExecutor) []ToolCallResponse {
	responses := make([]ToolCallResponse, len(tools))

	for i, tool := range tools {
		responses[i] = ExecuteTool(ctx, tool, executor)
	}

	return responses
}

// ExtractTools 从模型回复中提取工具调用
func ExtractTools(content string) ([]Tool, bool) {
	logger.Debug("开始提取工具调用", zap.Int("content_length", len(content)))

	// 查找工具调用的JSON格式 - 支持多种格式
	var jsonStr string

	// 尝试查找```json格式
	if start := strings.Index(content, "```json"); start != -1 {
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr = strings.TrimSpace(content[start+7 : start+end])
			logger.Debug("找到```json格式的JSON")
		}
	}

	// 如果没有找到```json，尝试查找普通的```格式
	if jsonStr == "" {
		if start := strings.Index(content, "```"); start != -1 {
			end := strings.Index(content[start:], "```")
			if end != -1 {
				jsonStr = strings.TrimSpace(content[start+3 : start+end])
				logger.Debug("找到```格式的JSON")
			}
		}
	}

	// 如果没有找到代码块，尝试直接查找JSON数组
	if jsonStr == "" {
		// 查找以[开头的JSON数组
		start := strings.Index(content, "[{")
		if start != -1 {
			// 找到匹配的结束括号
			bracketCount := 0
			end := start
			for i := start; i < len(content); i++ {
				if content[i] == '[' {
					bracketCount++
				} else if content[i] == ']' {
					bracketCount--
					if bracketCount == 0 {
						end = i + 1
						break
					}
				}
			}
			if end > start {
				jsonStr = strings.TrimSpace(content[start:end])
				logger.Info("找到JSON数组格式")
			}
		}
	}

	// 如果还是没有找到，尝试查找单个工具对象
	if jsonStr == "" {
		start := strings.Index(content, "{")
		if start != -1 {
			// 找到匹配的结束括号
			bracketCount := 0
			end := start
			for i := start; i < len(content); i++ {
				if content[i] == '{' {
					bracketCount++
				} else if content[i] == '}' {
					bracketCount--
					if bracketCount == 0 {
						end = i + 1
						break
					}
				}
			}
			if end > start {
				jsonStr = strings.TrimSpace(content[start:end])
				logger.Info("找到单个JSON对象格式")
			}
		}
	}

	if jsonStr == "" {
		logger.Info("未找到任何JSON格式的工具调用")
		return nil, false
	}

	logger.Info("提取到的JSON字符串", zap.String("json_string", jsonStr))

	// 解析JSON
	var tools []Tool
	err := json.Unmarshal([]byte(jsonStr), &tools)
	if err != nil {
		logger.Error("解析JSON数组失败，尝试解析单个工具", zap.Error(err))
		// 尝试解析单个工具
		var singleTool Tool
		err = json.Unmarshal([]byte(jsonStr), &singleTool)
		if err != nil {
			logger.Error("无法解析工具调用", zap.Error(err))
			logger.Info("原始JSON", zap.String("original_json", jsonStr))
			return nil, false
		}
		tools = []Tool{singleTool}
		logger.Info("成功解析单个工具")
	} else {
		logger.Info("成功解析工具数组", zap.Int("tool_count", len(tools)))
	}

	return tools, len(tools) > 0
}

// HandleToolCall 处理用户直接输入的工具调用命令
func HandleToolCall(ctx context.Context, command string, executor ToolExecutor) {
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
			logger.Error("tools 错误: 无法解析参数", zap.Error(err))
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
	response := ExecuteTool(ctx, tool, executor)

	// 打印结果
	logger.Debug("tools 工具结果", zap.String("result", response.Result))

	if response.Error != "" {
		logger.Error("tools 调用错误", zap.String("error", response.Error))
	}
}

// FormatToolResponses 格式化工具调用结果
func FormatToolResponses(responses []ToolCallResponse) string {
	result := "工具执行结果:\n"
	result += "==============\n\n"

	for i, resp := range responses {
		result += fmt.Sprintf("工具调用 %d:\n", i+1)
		result += fmt.Sprintf("状态: %s\n", func() string {
			if resp.Error != "" {
				return "失败"
			}
			return "成功"
		}())

		if resp.Error != "" {
			result += fmt.Sprintf("错误信息: %s\n", resp.Error)
		} else {
			result += fmt.Sprintf("执行结果:\n%s\n", resp.Result)
		}

		result += "\n"
	}

	return result
}
