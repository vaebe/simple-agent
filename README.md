# Simple Agent

这是一个基于智谱GLM-4.5模型的简单AI代理，可以处理文本对话并执行文件操作和Shell命令。

## 功能特点

- 基于智谱GLM-4.5大语言模型的对话能力
- 支持文件操作（列表、读取、写入）
- 支持执行Shell命令
- 工具调用功能，允许模型自主选择和使用工具

## 项目结构

```
.
├── main.go         # 主程序入口
├── agent.go        # 代理核心实现
├── message.go      # 消息结构和常量定义
├── mcp_client.go   # 文件操作客户端
└── go.mod          # Go模块定义
```

## 安装和使用

### 前提条件

- Go 1.16 或更高版本
- 智谱AI API密钥

### 安装

1. 克隆仓库

```bash
git clone https://github.com/yourusername/simple-agent.git
cd simple-agent
```

2. 安装依赖

```bash
go mod tidy
```

### 配置

设置智谱AI API密钥环境变量：

```bash
export ZHIPU_API_KEY=your_api_key_here
```

### 运行

```bash
go run .
```

## 使用方法

### 基本对话

直接输入文本与代理对话。

### 工具调用

代理可以自动识别需要使用工具的请求，并调用相应的工具。

### 手动工具调用

你也可以手动调用工具：

```
/tool file_operation list [目录路径]
/tool file_operation read [文件路径]
/tool file_operation write [文件路径] [内容]
/tool shell_command execute [命令]
```

## 示例

```
与GLM-4.5聊天 (使用'ctrl-c'退出)
你: 你好，请介绍一下你自己
GLM-4.5: 你好！我是由智谱AI开发的GLM-4.5大语言模型。我可以帮助你回答问题、提供信息、进行对话，以及使用各种工具来完成特定任务。

我的能力包括：
1. 回答各种知识性问题
2. 协助编写和审查文本内容
3. 使用文件操作工具来管理文件
4. 执行Shell命令来完成系统操作

有什么我可以帮助你的吗？

你: 请列出当前目录的内容
GLM-4.5: 我将为你列出当前目录的内容。

```json
[
  {
    "type": "file_operation",
    "name": "list",
    "args": {
      "path": "."
    },
    "thought": "用户想要查看当前目录的内容，我应该使用file_operation工具的list功能。"
  }
]
```

工具调用结果:
目录 /path/to/current/directory 的内容:
- agent.go (文件)
- go.mod (文件)
- go.sum (文件)
- main.go (文件)
- mcp_client.go (文件)
- message.go (文件)
- README.md (文件)

这是当前目录中的所有文件。你可以看到项目的主要Go源代码文件、Go模块文件以及README文档。需要我解释其中任何文件的用途吗？
```

## 许可证

MIT