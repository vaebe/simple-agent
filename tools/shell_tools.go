package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecuteShellCommand 执行Shell命令
func ExecuteShellCommand(tool Tool) ToolCallResponse {
	// 获取命令参数
	cmdStr, ok := tool.Args["command"].(string)
	if !ok {
		return ToolCallResponse{Error: "缺少命令参数"}
	}

	// 安全检查：禁止执行危险命令
	dangerousCommands := []string{"rm -rf", "sudo", "su", "chmod 777", "dd if=", "> /dev/"}
	for _, dangerous := range dangerousCommands {
		if strings.Contains(strings.ToLower(cmdStr), strings.ToLower(dangerous)) {
			return ToolCallResponse{Error: fmt.Sprintf("出于安全考虑，禁止执行命令: %s", cmdStr)}
		}
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	// 获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ToolCallResponse{Error: "命令执行超时"}
		}
		return ToolCallResponse{
			Result: string(output),
			Error:  err.Error(),
		}
	}

	return ToolCallResponse{Result: string(output)}
} 