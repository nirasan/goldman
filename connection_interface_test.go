package goldman_test

import (
	"github.com/nirasan/goldman"
	"golang.org/x/net/context"
	"testing"
)

type IConnectionTestRequest struct {
}

type IConnectionTestResponse struct {
	Success bool `json:"success"`
}

func IConnectionTest(conn goldman.IConnection, data *IConnectionTestRequest) {
	conn.Emit("iconnection_test", &IConnectionTestResponse{Success: true})
}

func TestIConnection_UseHttpTest(t *testing.T) {
	router := goldman.NewRouter()
	router.On("iconnection_test", IConnectionTest)

	ts, conn, err := createServerClient(router)
	fatalNotNil(t, err)
	defer ts.Close()
	defer conn.Close()

	err = writeMessage(conn, `iconnection_test {}`)
	fatalNotNil(t, err)

	ret, err := readMessage(conn)
	fatalNotNil(t, err)

	if ret != `iconnection_test {"success":true}` {
		t.Error("invalid success value: " + ret)
	}
}

type mockConn struct {
	Msg  string
	Data interface{}
}

func (conn *mockConn) Emit(msg string, data interface{})    { conn.Msg = msg; conn.Data = data }
func (conn *mockConn) Close()                               {}
func (conn *mockConn) GetContext() context.Context          { return nil }
func (conn *mockConn) SetContext(context.Context)           {}
func (conn *mockConn) GetRoomManager() *goldman.RoomManager { return nil }

func TestIConnection_UseMock(t *testing.T) {
	conn := &mockConn{}
	IConnectionTest(conn, &IConnectionTestRequest{})
	if conn.Msg != "iconnection_test" {
		t.Error("invalid message: " + conn.Msg)
	}
	if _, ok := conn.Data.(*IConnectionTestResponse); !ok {
		t.Error("data type is not *IConnectionTestResponse")
	}
}
