// Package websocket 提供 WebSocket 测试的辅助函数和处理器
package websocket

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// echoHandler 回显所有收到的消息
func echoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// 正常关闭或异常关闭
			break
		}

		// 回显消息
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
		}
	}
}

// jsonEchoHandler 回显 JSON 消息
func jsonEchoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var data map[string]interface{}
		err := conn.ReadJSON(&data)
		if err != nil {
			break
		}

		err = conn.WriteJSON(data)
		if err != nil {
			break
		}
	}
}

// pingPongHandler 处理 Ping/Pong
func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 设置 Pong 处理函数
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	// 定期发送 Ping
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := conn.WriteControl(websocket.PingMessage, []byte("ping from server"), time.Now().Add(5*time.Second))
				if err != nil {
					return
				}
			}
		}
	}()

	// 读取消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// blockingHandler 阻塞的处理器（用于测试写入超时）
func blockingHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 不读取消息，让写入缓冲区填满
	time.Sleep(10 * time.Second)
}

// broadcastHandler 广播消息给所有连接的客户端
func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 读取客户端消息并回显
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 回显给发送者
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
		}
	}
}
