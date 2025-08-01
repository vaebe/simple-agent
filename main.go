package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理，以便优雅地退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n接收到退出信号，正在退出...")
		cancel()
	}()

	// 从环境变量获取API密钥
	apiKey := os.Getenv("ZHIPU_API_KEY")
	if apiKey == "" {
		fmt.Println("错误: 未设置ZHIPU_API_KEY环境变量")
		fmt.Println("请设置环境变量: export ZHIPU_API_KEY=your_api_key")
		os.Exit(1)
	}

	// 创建代理配置
	config := AgentConfig{
		APIKey:       apiKey,
		SystemPrompt: DEFAULT_SYSTEM_PROMPT,
		Tools:        []string{TOOL_FILE_OPERATION, TOOL_SHELL_COMMAND},
	}

	// 创建代理实例
	agent := NewAdvancedAgent(config, getUserInput)

	// 显示欢迎信息
	fmt.Println("\n欢迎使用智谱AI GLM-4.5模型驱动的智能助手！")
	fmt.Println("该模型具有强大的推理能力、稳定的代码生成和多工具协同处理能力，同时运行速度更快。")
	fmt.Println("输入您的问题或指令，输入'exit'或'quit'退出。\n")

	// 运行代理
	err := agent.Run(ctx)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

// getUserInput 从标准输入获取用户输入
func getUserInput() (string, bool) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}

	// 去除输入末尾的换行符
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	// 检查是否是退出命令
	if input == "exit" || input == "quit" {
		return "", false
	}

	return input, true
}