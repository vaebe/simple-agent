package tools

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

// 定义工具类型常量
const (
	TOOL_FILE_OPERATION = "file_operation" // 文件操作工具
	TOOL_SHELL_COMMAND  = "shell_command"  // Shell命令工具
) 