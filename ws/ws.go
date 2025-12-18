package ws

import (
	"encoding/json"
	"fmt"
	"github.com/chihqiang/logx"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"time"
)

type ResponseCode int
type MessageType string

const (
	// OutputResponseCode 输出内容
	OutputResponseCode ResponseCode = 0
	// ConSuccessResponseCode 连接成功
	ConSuccessResponseCode ResponseCode = 1
	// ErrorResponseCode 错误信息
	ErrorResponseCode ResponseCode = 2
)

const (
	ConnectMessageTypeCode MessageType = "connect"
	DataMessageTypeCode    MessageType = "data"
	ResizeMessageTypeCode  MessageType = "resize"
)

type WSMessage struct {
	Type     MessageType `json:"type"`     // "connect", "data", "resize"
	Host     string      `json:"host"`     // SSH IP
	Port     int         `json:"port"`     // SSH Port
	Username string      `json:"username"` // 用户名
	Password string      `json:"password"` // 密码
	Data     string      `json:"data"`     // 输入命令
	Rows     int         `json:"rows"`     // PTY 高度
	Cols     int         `json:"cols"`     // PTY 宽度
}

type WSResponse struct {
	Code    ResponseCode `json:"code"`    // 0=输出,1=成功,2=错误
	Message string       `json:"message"` // 内容
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Helper: 发送结构化消息
func sendWSMessage(ws *websocket.Conn, code ResponseCode, msg string) {
	resp := WSResponse{
		Code:    code,
		Message: msg,
	}
	data, _ := json.Marshal(resp)
	_ = ws.WriteMessage(websocket.TextMessage, data)
}

// wsWriter: stdout/stderr 封装为 JSON
type wsWriter struct {
	ws *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	sendWSMessage(w.ws, OutputResponseCode, string(p))
	return len(p), nil
}

func SSHHandle(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Error("[WS] Upgrade error: %v", err)
		return
	}
	defer ws.Close()
	logx.Info("[WS] 新连接: %s", r.RemoteAddr)

	var (
		client  *ssh.Client
		session *ssh.Session
		stdin   io.WriteCloser
	)

	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			logx.Info("[WS] 连接关闭: %v", err)
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			logx.Warn("[WS] 消息解析失败: %v", err)
			sendWSMessage(ws, ErrorResponseCode, "消息解析失败\r\n")
			continue
		}
		logx.Info("[WS] 接收到消息: %+v", msg)
		switch msg.Type {
		case ConnectMessageTypeCode:
			port := msg.Port
			if port == 0 {
				port = 22
			}
			cfg := &ssh.ClientConfig{
				User:            msg.Username,
				Auth:            []ssh.AuthMethod{ssh.Password(msg.Password)},
				Timeout:         10 * time.Second,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			addr := fmt.Sprintf("%s:%d", msg.Host, port)
			logx.Info("[SSH] 尝试连接: %s", addr)

			client, err = ssh.Dial("tcp", addr, cfg)
			if err != nil {
				logx.Warn("[SSH] 连接失败: %v", err)
				sendWSMessage(ws, ErrorResponseCode, "SSH 连接失败\r\n")
				continue
			}
			logx.Info("[SSH] 连接成功: %s", addr)

			session, err = client.NewSession()
			if err != nil {
				logx.Warn("[SSH] 创建会话失败: %v", err)
				sendWSMessage(ws, ErrorResponseCode, "创建会话失败\r\n")
				continue
			}

			modes := ssh.TerminalModes{
				ssh.ECHO:          1,
				ssh.TTY_OP_ISPEED: 14400,
				ssh.TTY_OP_OSPEED: 14400,
			}
			if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
				logx.Warn("[SSH] 请求 PTY 失败: %v", err)
			}
			stdin, _ = session.StdinPipe()
			stdout, _ := session.StdoutPipe()
			stderr, _ := session.StderrPipe()

			go func() {
				if n, err := io.Copy(&wsWriter{ws}, stdout); err != nil {
					logx.Warn("[SSH] stdout 发送错误: %v", err)
				} else {
					logx.Info("[SSH] stdout 已发送 %d 字节", n)
				}
			}()
			go func() {
				if n, err := io.Copy(&wsWriter{ws}, stderr); err != nil {
					logx.Warn("[SSH] stderr 发送错误: %v", err)
				} else {
					logx.Info("[SSH] stderr 已发送 %d 字节", n)
				}
			}()

			if err := session.Shell(); err != nil {
				logx.Warn("[SSH] 启动 Shell 失败: %v", err)
				sendWSMessage(ws, ErrorResponseCode, "启动 Shell 失败\r\n")
				continue
			}

			sendWSMessage(ws, ConSuccessResponseCode, "SSH 连接成功\r\n")

		case ResizeMessageTypeCode:
			if session != nil {
				if err := session.WindowChange(msg.Rows, msg.Cols); err != nil {
					logx.Warn("[SSH] resize 错误: %v", err)
					sendWSMessage(ws, ErrorResponseCode, fmt.Sprintf("Resize 失败: %v\r\n", err))
				}
			}

		case DataMessageTypeCode:
			if stdin != nil {
				logx.Info("[SSH] 输入命令长度: %d 字节", len(msg.Data))
				if _, err := stdin.Write([]byte(msg.Data)); err != nil {
					logx.Warn("[SSH] 写入 stdin 失败: %v", err)
					sendWSMessage(ws, ErrorResponseCode, fmt.Sprintf("命令发送失败: %v\r\n", err))
				}
			}
		}
	}

	// 安全关闭资源
	if stdin != nil {
		_ = stdin.Close()
	}
	if session != nil {
		_ = session.Close()
	}
	if client != nil {
		_ = client.Close()
	}

	logx.Info("[WS] 连接资源已释放: %s", r.RemoteAddr)
}
