# 智慧学工自动签到工具

这是一个用于智慧学工（wisestu）的实习自动签到 Go 程序，支持通过 LLM API 自动识别验证码和定时执行。
<img width="1222" height="239" alt="IMG_20250812_205629" src="https://github.com/user-attachments/assets/44d5251d-4e58-4df5-a29e-d5ab5a4d49d1" />


## ✨ 功能特性

- **自动登录**：处理登录流程，包括验证码识别。
- **验证码识别**：支持接入 gpt-4.1-mini API，自动识别并计算图片验证码。
- **自动签到**：自动获取未签到任务并执行签到。
- **定时任务**：支持 Cron 表达式配置，实现定时自动签到。
- **失败重试**：内置网络请求和签到失败的重试机制。
- **灵活配置**：通过 YAML 配置文件或命令行参数进行配置。
- **结构化日志**：详细的日志记录，便于问题排查。

## 🚀 快速开始

### 1. 环境准备

- Go 1.22 或更高版本
- 一个能够访问 LLM API 的 Key

### 2. 下载和编译

```bash
# 克隆项目
git clone <your-repo-url>
cd zhxg-signin

# 下载依赖
go mod tidy

# 编译
go build -o zhxg-signin ./cmd/zhxg-signin
```

### 3. 配置

复制配置文件模板并填写您的信息：

```bash
cp configs/config.yaml.example configs/config.yaml
```

编辑 `configs/config.yaml` 文件，至少需要填写以下字段：

- `user.username`: 您的学号
- `user.password`: 您的密码
- `llm.api_key`: 您的 LLM API Key
- `location.longitude`: 签到时使用的经度
- `location.latitude`: 签到时使用的纬度

### 4. 运行

#### 执行一次性签到

您可以通过命令行参数覆盖配置文件中的设置：

```bash
./zhxg-signin run \
  --config ./configs/config.yaml \
  --username "你的学号" \
  --password "你的密码" \
  --api-key "你的LLM_API_KEY"
```

#### 启动定时服务

以守护进程模式运行，程序将根据配置文件中的 Cron 表达式定时执行签到：

```bash
./zhxg-signin daemon --config ./configs/config.yaml
```

## ⚙️ 配置说明

详细的配置选项请参考 `configs/config.yaml.example` 文件。

- **user**: 用户凭据。
- **location**: 签到时使用的地理位置坐标。
- **llm**: LLM API 相关配置。
- **signin**: 签到 API 和重试策略。
- **scheduler**: 定时任务配置。
- **logging**: 日志配置。

## 🤝 贡献

欢迎提交 Pull Request 或 Issue。

## 📄 许可证

本项目采用 MIT 许可证。
