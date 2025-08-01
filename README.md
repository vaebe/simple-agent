# Simple Agent

- 参考 <https://ampcode.com/how-to-build-an-agent> 实现
- 这是一个基于智谱AI GLM-4.5-Flash模型的智能助手，它不是很聪明又时候会卡住！

## 🚀 主要特性

- **GLM-4.5-Flash模型**: 使用智谱AI最新的GLM-4.5-Flash模型，具备128K上下文窗口
- **智能工具调用**: 支持文件操作和Shell命令执行
- **思考模式**: 启用动态思考模式，提供更深层次的推理分析
- **安全保护**: 内置安全检查机制，防止危险操作
- **流式输出**: 支持实时流式响应，提升用户交互体验

## 📋 功能列表

### 核心能力

- 智能问答和对话
- 复杂推理和问题解决
- 代码生成和调试
- 文件操作（列出、读取、写入）
- Shell命令执行（带安全检查）

### 工具支持

1. **文件操作工具**
   - `list`: 列出目录内容
   - `read`: 读取文件内容
   - `write`: 写入文件内容

2. **Shell命令工具**
   - `execute`: 执行Shell命令（带安全检查）

## 🔧 安装和配置

### 环境要求

- Go 1.16 或更高版本
- 智谱AI API密钥

### 安装步骤

1. 克隆项目

```bash
git clone git@github.com:vaebe/simple-agent.git
cd simple-agent
```

2. 安装依赖

```bash
go mod tidy
```

3. 设置API密钥

```bash
export ZHIPU_API_KEY=your_api_key_here
```

4. 运行程序

```bash
go run .
```

## 🛡️ 安全特性

### 文件操作安全

- 禁止访问上级目录（`..`）
- 禁止访问绝对路径（`/`开头）
- 文件读取大小限制（最大1MB）

### Shell命令安全

- 禁止执行危险命令（`rm -rf`、`sudo`、`su`等）
- 命令执行超时限制（30秒）
- 输入验证和清理

## 📖 使用示例

### 基本对话

```bash
你: 你好，请介绍一下你自己
GLM-4.5-Flash: 你好！我是基于智谱AI GLM-4.5-Flash模型的智能助手...
```

### 文件操作

```bash
你: 请列出当前目录的内容
GLM-4.5-Flash: 我将使用文件操作工具来列出当前目录的内容。

[工具调用]
{
  "type": "file_operation",
  "name": "list",
  "args": {
    "path": "."
  }
}
```

## 🔗 相关链接

- [How to build an agent](https://ampcode.com/how-to-build-an-agent)
- [智谱AI GLM-4.5-Flash文档](https://docs.bigmodel.cn/cn/guide/models/free/glm-4.5-flash)
- [智谱AI智能体API文档](https://docs.bigmodel.cn/api-reference/agent-api/%E6%99%BA%E8%83%BD%E4%BD%93%E5%AF%B9%E8%AF%9D)