package websocket

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket_ReadDeadline 测试读取超时
func TestWebSocket_ReadDeadline(t *testing.T) {
	t.Run("读取超时", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 设置很短的读取超时
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		// 服务器不会发送消息，不会有ping-ping交互，所以读取会超时
		_, _, err = conn.ReadMessage()
		if err == nil {
			t.Error("期望读取超时，但读取成功")
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Log("成功触发读取超时")
		} else {
			t.Errorf("读取错误（可能是超时）: %v", err)
		}
	})
}

// TestWebSocket_WriteDeadline 测试写入超时
func TestWebSocket_WriteDeadline(t *testing.T) {
	t.Run("写入超时", func(t *testing.T) {
		// 创建一个会阻塞的服务器
		server := httptest.NewServer(http.HandlerFunc(blockingHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 设置很短的写入超时
		conn.SetWriteDeadline(time.Now().Add(100 * time.Microsecond))

		// 尝试写入大量数据（可能会阻塞）
		largeData := make([]byte, 10*1024*1024) // 10MB
		err = conn.WriteMessage(websocket.BinaryMessage, largeData)
		if err == nil {
			t.Errorf("写入成功（可能未触发超时）")
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Log("成功触发写入超时")
		} else {
			t.Errorf("写入错误: %v", err)
		}
	})
}
