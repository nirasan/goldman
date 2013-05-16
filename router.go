package golem

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Router struct {
	callbacks map[string]func(*Connection, []byte)
}

func NewRouter() *Router {
	hub.run()
	return &Router{
		callbacks: make(map[string]func(*Connection, []byte)),
	}
}

func (router *Router) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Origin not allowed", 403)
			return
		}

		socket, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)

		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		conn := &Connection{
			socket: socket,
			router: router,
			out:    make(chan []byte, outChannelSize),
		}

		hub.register <- conn
		go conn.writePump()
		conn.readPump()
	}
}

func (router *Router) On(name string, callback interface{}) {

	callbackDataType := reflect.TypeOf(callback).In(1)

	fmt.Println(callbackDataType, reflect.TypeOf([]byte{}))
	if reflect.TypeOf([]byte{}) == callbackDataType {
		router.callbacks[name] = callback.(func(*Connection, []byte))
		return
	}

	callbackValue := reflect.ValueOf(callback)
	callbackDataElem := callbackDataType.Elem()

	preCallbackParser := func(conn *Connection, data []byte) {
		result := reflect.New(callbackDataElem)

		err := json.Unmarshal(data, &result)
		if err == nil {
			args := []reflect.Value{reflect.ValueOf(conn), result}
			callbackValue.Call(args)
		} else {
			fmt.Println("[JSON-FORWARD]", data, err) // TODO: Proper debug output!
		}
	}
	router.callbacks[name] = preCallbackParser
}

func (router *Router) parse(conn *Connection, rawdata []byte) {
	rawstring := string(rawdata)
	data := strings.SplitN(rawstring, " ", 2)
	if len(data) == 2 {
		if callback, ok := router.callbacks[data[0]]; ok {
			callback(conn, []byte(data[1]))
		}
	}

	defer recover()
}