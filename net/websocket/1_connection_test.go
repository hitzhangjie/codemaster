package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

// TestWebSocket_Connection 测试基本的连接建立
func TestWebSocket_Connection(t *testing.T) {
	t.Run("建立连接", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Errorf("期望状态码 101，实际: %d", resp.StatusCode)
		}

		t.Logf("WebSocket 连接成功建立，状态码: %d", resp.StatusCode)
	})
}
