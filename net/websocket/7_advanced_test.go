package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

// TestWebSocket_MessageSizeLimit 测试消息大小限制
func TestWebSocket_MessageSizeLimit(t *testing.T) {
	t.Run("消息大小限制", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// 设置最大消息大小
			conn.SetReadLimit(512) // 512 bytes

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseMessageTooBig) {
						t.Logf("消息过大，连接关闭: %v", err)
					}
					return
				}
			}
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 发送超过限制的消息
		largeMessage := make([]byte, 1024) // 超过 512 bytes 限制
		err = conn.WriteMessage(websocket.TextMessage, largeMessage)
		if err != nil {
			t.Logf("发送大消息失败（预期）: %v", err)
		}

		// 尝试读取，应该收到关闭消息
		_, _, err = conn.ReadMessage()
		if err != nil {
			t.Logf("连接已关闭（预期）: %v", err)
		}
	})
}

// TestWebSocket_Subprotocol 测试子协议
func TestWebSocket_Subprotocol(t *testing.T) {
	t.Run("子协议协商", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			Subprotocols: []string{"chat", "superchat"},
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// 发送子协议信息
			conn.WriteMessage(websocket.TextMessage, []byte(conn.Subprotocol()))
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		dialer := websocket.Dialer{
			Subprotocols: []string{"chat"},
		}

		conn, resp, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 检查响应头中的子协议
		subprotocol := resp.Header.Get("Sec-WebSocket-Protocol")
		if subprotocol != "chat" {
			t.Errorf("期望子协议 'chat'，实际: %s", subprotocol)
		}

		// 读取服务器发送的子协议信息
		_, received, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("读取失败: %v", err)
		}

		if string(received) != "chat" {
			t.Errorf("期望子协议 'chat'，实际: %s", string(received))
		}

		t.Logf("子协议协商成功: %s", subprotocol)
	})
}

// TestWebSocket_OriginCheck 测试 Origin 检查
func TestWebSocket_OriginCheck(t *testing.T) {
	t.Run("Origin检查", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 只允许特定 origin
				return r.Header.Get("Origin") == "http://example.com"
			},
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				http.Error(w, "Upgrade failed", http.StatusBadRequest)
				return
			}
			defer conn.Close()

			conn.WriteMessage(websocket.TextMessage, []byte("OK"))
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"

		// 测试允许的 origin
		header := http.Header{}
		header.Set("Origin", "http://example.com")
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, header)
		if err != nil {
			t.Fatalf("连接失败（应该成功）: %v", err)
		}
		conn.Close()

		// 测试不允许的 origin
		header.Set("Origin", "http://evil.com")
		conn, _, err = dialer.Dial(wsURL, header)
		if err == nil {
			conn.Close()
			t.Error("期望连接失败（origin 不匹配），但连接成功")
		} else {
			t.Logf("Origin 检查成功，拒绝了不匹配的 origin: %v", err)
		}
	})
}
