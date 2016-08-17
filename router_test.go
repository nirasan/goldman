package golem

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Msg string `json:"msg"`
}

func Hello(conn *Connection, data *HelloRequest) {
	conn.Emit("hello", &HelloResponse{Msg: "hello " + data.Name})
}

func TestRouter_Handler(t *testing.T) {
	myrouter := NewRouter()
	myrouter.On("hello", Hello)

	ts, conn := createServerClient(t, myrouter)
	defer ts.Close()
	defer conn.Close()

	writeMessage(t, conn, `hello {"name":"goldman"}`)
	res := readMessage(t, conn)
	if res != `hello {"msg":"hello goldman"}` {
		t.Error("invalid message: " + res)
	}
}

func createServerClient(t *testing.T, router *Router) (*httptest.Server, *websocket.Conn) {
	ts := httptest.NewServer(http.HandlerFunc(router.Handler()))

	dialer := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	url := strings.Replace(ts.URL, "http://", "ws://", 1)
	header := http.Header{"Accept-Encoding": []string{"gzip"}}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		t.Fatal(err)
	}

	return ts, conn
}

func writeMessage(t *testing.T, conn *websocket.Conn, message string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		t.Fatal(err)
	}
}

func readMessage(t *testing.T, conn *websocket.Conn) string {
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if messageType != websocket.TextMessage {
		t.Error("invalid message type")
	}
	return string(p)
}
