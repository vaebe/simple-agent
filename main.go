package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"simple-agent/logger"
	"simple-agent/tools"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"go.uber.org/zap"
)

// 全局readline实例
var rl *readline.Instance

func main() {
	// 初始化日志系统
	logger.Init()
	defer logger.Sync()

	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理，以便优雅地退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("接收到退出信号，正在退出...")
		if rl != nil {
			rl.Close()
		}
		cancel()
	}()

	// 初始化readline实例
	var err error
	rl, err = readline.New("")
	if err != nil {
		logger.Fatal("初始化readline失败", zap.Error(err))
		os.Exit(1)
	}
	defer rl.Close()

	// 从环境变量获取API密钥
	apiKey := os.Getenv("ZHIPU_API_KEY")
	if apiKey == "" {
		logger.Error("未设置 ZHIPU_API_KEY 环境变量, 请设置环境变量: export ZHIPU_API_KEY=your_api_key")
		os.Exit(1)
	}

	// 创建代理配置
	config := AgentConfig{
		APIKey:       apiKey,
		SystemPrompt: DEFAULT_SYSTEM_PROMPT,
		Tools:        []string{tools.TOOL_FILE_OPERATION, tools.TOOL_SHELL_COMMAND},
	}

	// 创建代理实例
	agent := NewAdvancedAgent(config, getUserInput)

	// 显示欢迎信息
	fmt.Println("\033[32m \n欢迎使用 VCode 智能助手， 输入您的问题或指令，输入'exit'或'quit'退出。 \033[0m")

	// 运行代理
	err = agent.Run(ctx)
	if err != nil {
		logger.Error("程序运行出错", zap.Error(err))
		os.Exit(1)
	}
}

// getUserInput 从标准输入获取用户输入
func getUserInput() (string, bool) {
	if rl == nil {
		logger.Debug("readline实例未初始化")
		return "", false
	}

	// 设置提示符
	rl.SetPrompt("\u001b[94m你\u001b[0m: ")

	// 读取输入
	input, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			// Ctrl+C 退出
			return "", false
		}
		if err == io.EOF {
			// EOF (Ctrl+D) 退出
			return "", false
		}
		logger.Error("读取输入失败", zap.Error(err))

		return "", false
	}

	// 去除输入末尾的换行符和空格
	input = strings.TrimSpace(input)

	// 检查是否是退出命令
	if input == "exit" || input == "quit" {
		return "", false
	}

	// 检查空输入
	if input == "" {
		return "", true // 返回空字符串但继续循环
	}

	return input, true
}
