package golem

import (
	"fmt"
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
	ts := httptest.NewServer(http.HandlerFunc(myrouter.Handler()))
	defer ts.Close()

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
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`hello {"name":"goldman"}`))
	if err != nil {
		t.Fatal(err)
	}

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if messageType != websocket.TextMessage {
		t.Error("invalid message type")
	}
	if string(p) != `hello {"msg":"hello goldman"}` {
		t.Error(fmt.Sprintf("invalid message: %s", string(p)))
	}
}
