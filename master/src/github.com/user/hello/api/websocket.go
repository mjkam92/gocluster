package api

import (
	"github.com/googollee/go-socket.io"
	"log"
)

type SockMsg struct {
	Cmd      string
	Sock     socketio.Socket
	NodeList []Node
}

func delSocket(sockList []socketio.Socket, delSock socketio.Socket) []socketio.Socket {
	delIndex := -1

	//sList := sockList
	for i, sock := range sockList {
		if sock == delSock {
			delIndex = i
		}
	}

	if delIndex != -1 {
		sockList = append(sockList[:delIndex], sockList[delIndex+1:]...)
	}

	return sockList
}

func webSocketManager(sockMsgChan <-chan SockMsg) {
	webSocketList := []socketio.Socket{}

	for {
		select {
		case sockMsg := <-sockMsgChan:
			webSocketList = webSocketMsgHandler(sockMsg, webSocketList)
		}
	}
}

func webSocketMsgHandler(sockMsg SockMsg, webSocketList []socketio.Socket) []socketio.Socket {
	if sockMsg.Cmd == "ADD" {
		webSocketList = append(webSocketList, sockMsg.Sock)
	} else if sockMsg.Cmd == "DEL" {
		webSocketList = delSocket(webSocketList, sockMsg.Sock)
	} else if sockMsg.Cmd == "BROADCAST" {
		broadcast(webSocketList, sockMsg.NodeList)
	} else {
		log.Println("INVALID CMD")
	}
	log.Printf("WebSocket Msg Handler call. Print WebSocketList -> %v", webSocketList)
	return webSocketList
}

func getSocketIOServer(msgChan chan<- SockMsg) *socketio.Server {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")

		msg := SockMsg{Cmd: "ADD", Sock: so}
		msgChan <- msg

		so.On("disconnection", func() {
			log.Println("websocket disconnected")

			msg := SockMsg{Cmd: "DEL", Sock: so}
			msgChan <- msg
		})

	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	return server
}

func broadcast(sockList []socketio.Socket, nodeList []Node) {
	for _, sock := range sockList {
		sock.Emit("nodeListData", nodeList)
	}
}
