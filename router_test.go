package goldman

import (
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	router := NewRouter()
	router.On("hello", Hello)

	ts, conn, err := createServerClient(router)
	fatalNotNil(t, err)

	defer ts.Close()
	defer conn.Close()

	err = writeMessage(conn, `hello {"name":"goldman"}`)
	fatalNotNil(t, err)

	res, err := readMessage(conn)
	fatalNotNil(t, err)
	if res != `hello {"msg":"hello goldman"}` {
		t.Error("invalid message: " + res)
	}
}

type RequestPlus struct {
	Num1 int `json:"num1"`
	Num2 int `json:"num2"`
}

type ResponsePlus struct {
	Num int `json:"num"`
}

type RequestMinus struct {
	Num1 int `json:"num1"`
	Num2 int `json:"num2"`
}

type ResponseMinus struct {
	Num int `json:"num"`
}

func Plus(conn *Connection, data *RequestPlus) {
	conn.Emit("plus", &ResponsePlus{Num: data.Num1 + data.Num2})
}

func Minus(conn *Connection, data *RequestMinus) {
	conn.Emit("minus", &ResponseMinus{Num: data.Num1 - data.Num2})
}

func TestRouter_On(t *testing.T) {
	router := NewRouter()
	router.On("plus", Plus)
	router.On("minus", Minus)

	ts, conn, err := createServerClient(router)
	fatalNotNil(t, err)

	defer ts.Close()
	defer conn.Close()

	err = writeMessage(conn, `plus {"num1":1,"num2":2}`)
	fatalNotNil(t, err)

	res, err := readMessage(conn)
	fatalNotNil(t, err)

	if res != `plus {"num":3}` {
		t.Error("invalid message: " + res)
	}

	err = writeMessage(conn, `minus {"num1":1,"num2":2}`)
	fatalNotNil(t, err)

	res, err = readMessage(conn)
	fatalNotNil(t, err)

	if res != `minus {"num":-1}` {
		t.Error("invalid message: " + res)
	}

	err = writeMessage(conn, `hello {"name":"goldman"}`)
	fatalNotNil(t, err)

	_, err = readMessage(conn)
	if err == nil {
		t.Error("not registered callback executed")
	}
}

func createServerClient(router *Router) (*httptest.Server, *websocket.Conn, error) {
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
		return nil, nil, err
	}

	return ts, conn, nil
}

func writeMessage(conn *websocket.Conn, message string) error {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

func readMessage(conn *websocket.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(500000000 * time.Nanosecond))
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	if messageType != websocket.TextMessage {
		return "", errors.New("invalid message type")
	}
	return string(p), nil
}

func fatalNotNil(t *testing.T, e error) {
	if e != nil {
		t.Fatal(e)
	}
}
