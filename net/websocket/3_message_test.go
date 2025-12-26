package websocket

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket_TextMessage 测试文本消息的发送和接收
func TestWebSocket_TextMessage(t *testing.T) {
	t.Run("发送接收文本消息", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 发送文本消息
		message := "Hello, WebSocket!"
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			t.Fatalf("发送消息失败: %v", err)
		}

		// 接收消息
		_, received, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("接收消息失败: %v", err)
		}

		if string(received) != message {
			t.Errorf("期望消息: %s, 实际: %s", message, string(received))
		}

		t.Logf("成功发送和接收文本消息: %s", message)
	})

	t.Run("发送接收JSON消息", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(jsonEchoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 发送 JSON 消息
		data := map[string]interface{}{
			"type":      "message",
			"content":   "Hello from JSON",
			"timestamp": time.Now().Unix(),
		}
		err = conn.WriteJSON(data)
		if err != nil {
			t.Fatalf("发送 JSON 失败: %v", err)
		}

		// 接收 JSON 消息
		var received map[string]interface{}
		err = conn.ReadJSON(&received)
		if err != nil {
			t.Fatalf("接收 JSON 失败: %v", err)
		}

		if received["content"] != data["content"] {
			t.Errorf("期望 content: %v, 实际: %v", data["content"], received["content"])
		}

		t.Logf("成功发送和接收 JSON 消息: %+v", received)
	})
}

// TestWebSocket_BinaryMessage 测试二进制消息的发送和接收
func TestWebSocket_BinaryMessage(t *testing.T) {
	t.Run("发送接收二进制消息", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 发送二进制消息
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}
		err = conn.WriteMessage(websocket.BinaryMessage, binaryData)
		if err != nil {
			t.Fatalf("发送二进制消息失败: %v", err)
		}

		// 接收消息
		messageType, received, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("接收消息失败: %v", err)
		}

		if messageType != websocket.BinaryMessage {
			t.Errorf("期望消息类型 BinaryMessage，实际: %d", messageType)
		}

		if !bytes.Equal(received, binaryData) {
			t.Errorf("二进制数据不匹配")
		}

		t.Logf("成功发送和接收二进制消息，长度: %d", len(received))
	})

	t.Run("发送大二进制消息", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		// 发送 1MB 的二进制数据
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		err = conn.WriteMessage(websocket.BinaryMessage, largeData)
		if err != nil {
			t.Fatalf("发送大消息失败: %v", err)
		}

		_, received, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("接收大消息失败: %v", err)
		}

		if len(received) != len(largeData) {
			t.Errorf("期望长度: %d, 实际: %d", len(largeData), len(received))
		}

		t.Logf("成功发送和接收大二进制消息，大小: %d bytes", len(received))
	})
}
