package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket_ConcurrentReadWrite 测试并发读写
func TestWebSocket_ConcurrentReadWrite(t *testing.T) {
	t.Run("并发读写", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}
		defer conn.Close()

		var wg sync.WaitGroup
		messageCount := 100

		// 并发写入
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < messageCount; i++ {
				msg := fmt.Sprintf("消息 %d", i)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					t.Errorf("写入失败: %v", err)
					return
				}
			}
		}()

		// 并发读取
		wg.Add(1)
		go func() {
			defer wg.Done()
			receivedCount := 0
			for receivedCount < messageCount {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						t.Errorf("读取失败: %v", err)
					}
					return
				}
				receivedCount++
			}
			t.Logf("成功接收 %d 条消息", receivedCount)
		}()

		wg.Wait()
		t.Log("并发读写测试完成")
	})
}

// TestWebSocket_MultipleClients 测试多客户端连接
func TestWebSocket_MultipleClients(t *testing.T) {
	t.Run("多客户端连接", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(broadcastHandler))
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/ws"
		clientCount := 5
		var wg sync.WaitGroup

		for i := 0; i < clientCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err != nil {
					t.Errorf("客户端 %d 连接失败: %v", id, err)
					return
				}
				defer conn.Close()

				time.Sleep(2 * time.Second)

				// 发送消息
				msg := fmt.Sprintf("来自客户端 %d", id)
				err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					t.Errorf("客户端 %d 发送失败: %v", id, err)
					return
				}

				// 接收消息
				for j := 0; j < clientCount; j++ {
					conn.SetReadDeadline(time.Now().Add(time.Second))
					_, received, err := conn.ReadMessage()
					if err != nil {
						if websocket.IsUnexpectedCloseError(err) {
							t.Errorf("客户端 %d 接收失败: %v", id, err)
						}
						return
					}
					t.Logf("客户端 %d 收到消息: %s", id, string(received))
				}
			}(i)
		}

		wg.Wait()
		t.Logf("多客户端测试完成，共 %d 个客户端", clientCount)
	})
}
