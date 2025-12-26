package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket_PingPong 测试 Ping/Pong 心跳机制
func TestWebSocket_PingPong(t *testing.T) {
	t.Run("服务器发送Ping客户端自动回复Pong", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(pingPongHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// 在后台 goroutine 中开始读取，这样才能处理控制消息（如 Ping）
		// gorilla/websocket 库在 ReadMessage() 过程中会自动处理控制消息
		// 当服务器发送 Ping 时，客户端会自动回复 Pong（这是自动的，不需要设置处理函数）
		readErr := make(chan error, 1)
		readOK := make(chan bool, 1)
		go func() {
			conn.SetPingHandler(func(appData string) error {
				t.Logf("收到 Ping: %s", appData)
				readOK <- true
				return nil
			})
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					readErr <- err
					return
				}
				// 处理普通消息（TextMessage, BinaryMessage）
				//
				// 注意：Ping 消息会在 ReadMessage() 内部自动处理，客户端会自动回复 Pong
				// 但 ReadMessage() 不会返回 Ping 消息类型，所以我们需要通过其他方式验证
			}
		}()

		// 等待一段时间，让服务器发送 Ping（服务器每 1 秒发送一次）
		// 客户端会自动回复 Pong，连接保持活跃
		select {
		case err := <-readErr:
			t.Errorf("读取消息出错: %v", err)
		case <-readOK:
			// 2 秒内没有错误，说明连接正常，Ping/Pong 机制工作正常
			t.Log("连接正常，Ping/Pong 机制工作正常")
		}
	})

	t.Run("客户端发送Ping服务器自动回复Pong", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(pingPongHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 设置 Pong 处理函数，用于接收服务器回复的 Pong
		pongReceived := make(chan bool, 1)
		conn.SetPongHandler(func(appData string) error {
			t.Logf("收到服务器回复的 Pong: %s", appData)
			pongReceived <- true
			return nil
		})

		// 在后台 goroutine 中开始读取，这样才能处理控制消息（如 Pong）
		readErr := make(chan error, 1)
		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					readErr <- err
					return
				}
			}
		}()

		// 客户端发送 Ping
		err = conn.WriteControl(websocket.PingMessage, []byte("ping from client"), time.Now().Add(5*time.Second))
		if err != nil {
			t.Fatalf("发送 Ping 失败: %v", err)
		}

		t.Log("客户端成功发送 Ping")

		// 等待服务器自动回复 Pong
		select {
		case <-pongReceived:
			t.Log("成功收到服务器回复的 Pong")
		case err := <-readErr:
			t.Errorf("读取消息出错: %v", err)
		case <-time.After(2 * time.Second):
			t.Error("未在预期时间内收到服务器回复的 Pong")
		}
	})

	t.Run("服务器发送Pong客户端接收", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// 服务器发送 Pong 消息（通常这是对客户端 Ping 的回复，但也可以主动发送）
			err = conn.WriteControl(websocket.PongMessage, []byte("pong from server"), time.Now().Add(5*time.Second))
			if err != nil {
				return
			}

			// 保持连接一段时间
			time.Sleep(1 * time.Second)
		}))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 设置 Pong 处理函数
		pongReceived := make(chan bool, 1)
		conn.SetPongHandler(func(appData string) error {
			t.Logf("收到 Pong: %s", appData)
			pongReceived <- true
			return nil
		})

		// 在后台 goroutine 中开始读取，这样才能处理控制消息（如 Pong）
		readErr := make(chan error, 1)
		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					readErr <- err
					return
				}
			}
		}()

		// 等待服务器发送 Pong
		select {
		case <-pongReceived:
			t.Log("成功收到 Pong 响应")
		case err := <-readErr:
			t.Errorf("读取消息出错: %v", err)
		case <-time.After(2 * time.Second):
			t.Error("未在预期时间内收到 Pong")
		}
	})

	t.Run("显式开始读取示例", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(pingPongHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 方法1: 在 goroutine 中持续读取（推荐用于生产环境）
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					t.Logf("读取结束: %v", err)
					return
				}
				// 处理普通消息（TextMessage, BinaryMessage）
				if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
					t.Logf("收到消息: %s", string(message))
				}
				// 控制消息（Ping/Pong/Close）会在 ReadMessage() 内部自动处理
			}
		}()

		// 等待一段时间，让读取循环运行
		time.Sleep(2 * time.Second)

		// 关闭连接，结束读取循环
		conn.Close()
		<-done
		t.Log("读取循环已结束")
	})
}
