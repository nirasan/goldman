package goldman_test

import (
	"github.com/nirasan/goldman"
	"golang.org/x/net/context"
	"testing"
)

var ctxKey = 1

type CtxTestResponse struct {
	Value int `json:"value"`
}

func CtxTest1(conn *goldman.Connection) {
	ctx := conn.GetContext()
	ctx = context.WithValue(ctx, ctxKey, 100)
	conn.SetContext(ctx)
}

func CtxTest2(conn *goldman.Connection) {
	ctx := conn.GetContext()
	v, _ := ctx.Value(ctxKey).(int)
	conn.Emit("ctx_test", &CtxTestResponse{Value: v})
}

func TestConnection_GetContext(t *testing.T) {
	router := goldman.NewRouter()
	router.On("ctx_test1", CtxTest1)
	router.On("ctx_test2", CtxTest2)

	ts, conn, err := createServerClient(router)
	fatalNotNil(t, err)

	defer ts.Close()
	defer conn.Close()

	err = writeMessage(conn, `ctx_test1 {}`)
	fatalNotNil(t, err)

	err = writeMessage(conn, `ctx_test2 {}`)
	fatalNotNil(t, err)

	ret, err := readMessage(conn)
	fatalNotNil(t, err)

	if ret != `ctx_test {"value":100}` {
		t.Error("invalid ctx value: " + ret)
	}
}

type GetRoomManagerResponse struct {
	Success bool `json:"success"`
}

func GetRoomManager(conn *goldman.Connection) {
	room_manager := conn.GetRoomManager()
	if room_manager != nil {
		conn.Emit("room_manager", &GetRoomManagerResponse{Success: true})
	}
}

func TestConnection_GetRoomManager(t *testing.T) {
	router := goldman.NewRouter()
	router.On("room_manager", GetRoomManager)

	ts, conn, err := createServerClient(router)
	fatalNotNil(t, err)
	defer ts.Close()
	defer conn.Close()

	err = writeMessage(conn, `room_manager {}`)
	fatalNotNil(t, err)

	ret, err := readMessage(conn)
	fatalNotNil(t, err)

	if ret != `room_manager {"success":true}` {
		t.Error("invalid success value: " + ret)
	}
}
