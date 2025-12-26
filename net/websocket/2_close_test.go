package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket_Close 测试连接关闭
func TestWebSocket_Close(t *testing.T) {
	t.Run("正常关闭连接", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 正常关闭
		closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "正常关闭")
		err = conn.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(5*time.Second))
		if err != nil {
			t.Fatalf("发送关闭消息失败: %v", err)
		}

		// 尝试读取，应该收到关闭消息
		_, _, err = conn.ReadMessage()
		if err == nil {
			t.Error("期望读取失败（连接已关闭），但读取成功")
		} else if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			t.Logf("关闭错误: %v", err)
		}

		t.Log("正常关闭连接成功")
	})

	t.Run("异常关闭连接", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 直接关闭连接（不发送关闭消息）
		err = conn.Close()
		if err != nil {
			t.Fatalf("关闭连接失败: %v", err)
		}

		// 尝试写入，应该失败
		err = conn.WriteMessage(websocket.TextMessage, []byte("test"))
		if err == nil {
			t.Error("期望写入失败（连接已关闭），但写入成功")
		}

		t.Log("异常关闭连接成功")
	})
}
