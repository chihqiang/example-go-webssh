# example-go-webssh

webssh-go 是一个基于 Go + WebSocket + SSH 实现的轻量级 Web SSH 终端系统。通过浏览器即可安全地连接远程服务器，获得与本地终端一致的真实交互体验，无需安装任何客户端。

## 功能特点

- 🚀 轻量级设计，部署简单
- 🌐 浏览器访问，跨平台兼容
- 💻 完整的终端体验，支持命令行操作
- 📱 响应式设计，适配不同屏幕尺寸
- 🔒 支持 SSH 密码认证
- 📏 支持终端窗口大小调整
- 📝 实时日志记录

## 技术栈

- **后端**: Go 语言
- **前端**: HTML5 + JavaScript
- **通信**: WebSocket
- **终端**: xterm.js
- **SSH**: golang.org/x/crypto/ssh

## 快速开始

### 安装依赖

```bash
go mod download
```

### 运行项目

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 构建可执行文件

```bash
go build -o webssh main.go
```

然后运行:

```bash
./webssh
```

## 使用方法

1. 在浏览器中访问 `http://localhost:8080`
2. 填写服务器信息:
   - **服务器 IP**: 远程服务器的 IP 地址
   - **端口**: SSH 服务端口 (默认: 22)
   - **用户名**: SSH 登录用户名 (默认: root)
   - **密码**: SSH 登录密码
3. 点击 "连接" 按钮
4. 在终端窗口中输入命令，开始操作远程服务器

## 项目结构

```
example-webssh-go/
├── logx/              # 日志模块
│   ├── formatter.go   # 日志格式化
│   ├── level.go       # 日志级别
│   ├── log.go         # 日志接口
│   └── logger.go      # 日志实现
├── static/            # 静态资源
│   ├── favicon.ico    # 网站图标
│   └── index.html     # 前端页面
├── ws/                # WebSocket 模块
│   └── ws.go          # WebSocket 和 SSH 连接实现
├── .gitignore         # Git 忽略文件
├── LICENSE            # 许可证文件
├── README.md          # 项目说明文档
├── go.mod             # Go 模块依赖
├── go.sum             # Go 模块校验
└── main.go            # 主程序入口
```

## 核心功能实现

### WebSocket 连接

- 前端通过 WebSocket 与后端建立实时通信
- 支持三种消息类型:
  - `connect`: 建立 SSH 连接
  - `data`: 传输终端输入数据
  - `resize`: 调整终端窗口大小

### SSH 连接管理

- 使用 `golang.org/x/crypto/ssh` 库实现 SSH 连接
- 支持密码认证
- 自动处理终端窗口大小调整
- 安全关闭连接资源

### 终端模拟

- 使用 xterm.js 实现浏览器端终端
- 支持完整的终端功能
- 响应式设计，自动适应窗口大小

## 配置说明

### 服务器配置

在 `main.go` 文件中可以修改以下配置:

```go
const (
    serverAddr   = ":8080"      // 服务器监听地址
    readTimeout  = 15 * time.Second  // 读取超时
    writeTimeout = 15 * time.Second  // 写入超时
    idleTimeout  = 60 * time.Second  // 空闲超时
)
```

### WebSocket 配置

在 `ws/ws.go` 文件中可以修改 WebSocket 配置:

```go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,   // 读取缓冲区大小
    WriteBufferSize: 1024,   // 写入缓冲区大小
    CheckOrigin: func(r *http.Request) bool {
        return true          // 允许所有来源的请求
    },
}
```

## 注意事项

1. **安全性**: 本项目目前仅支持密码认证，建议在生产环境中使用更安全的认证方式
2. **防火墙**: 确保服务器的 8080 端口已开放
3. **SSH 服务**: 确保远程服务器已启用 SSH 服务
4. **浏览器兼容性**: 建议使用现代浏览器（Chrome、Firefox、Safari、Edge）
5. **生产环境**: 生产环境中建议配置 HTTPS 以加密通信

