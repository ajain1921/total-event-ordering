package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type MPNode struct {
	identifier string
	hostname   string
	port       string
}

type ConnectionStatus struct {
	outbound bool
	inbound  bool
}

type Message struct {
	node    string
	content string
}

type BasicMulticast struct {
	currentNode *MPNode
	otherNodes  []*MPNode
	connections map[string]*ConnectionStatus
	writer      chan Message
	receiver    chan string
}

func (multicast *BasicMulticast) Setup() error {
	multicast.receiver = make(chan string)
	multicast.connections = make(map[string]*ConnectionStatus)

	ln, err := net.Listen("tcp", ":"+multicast.currentNode.port)
	if err != nil {
		return err
	}

	for _, node := range multicast.otherNodes {
		multicast.connections[node.identifier] = &ConnectionStatus{}
		go multicast.connect(node)
	}

	for range multicast.otherNodes {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go multicast.handleConnection(conn)
	}

	return nil
}

func (multicast *BasicMulticast) Receiver() <-chan string {
	return multicast.receiver
}

func (multicast *BasicMulticast) ready() bool {
	for _, node := range multicast.otherNodes {
		connectionStatus := multicast.connections[node.identifier]
		if !connectionStatus.inbound || !connectionStatus.outbound {
			return false
		}
	}
	return true
}

func (multicast *BasicMulticast) sendMessage(conn net.Conn, message Message) {
	// TODO: Impl
	fmt.Fprintln(conn, message.node+" "+message.content)
}

func (multicast *BasicMulticast) parseMessage(line string) Message {
	var identifier string
	fmt.Sscanf(line, "%s connected", &identifier)
	return Message{node: identifier, content: "connected"}
}

func (multicast *BasicMulticast) connect(node *MPNode) {
	conn, err := net.Dial("tcp", node.hostname+":"+node.port)
	if err != nil {
		fmt.Println("trying to connect to " + node.identifier + " but failed")
		return
	}
	defer conn.Close()

	connectionMessage := Message{node: multicast.currentNode.identifier, content: "connected"}
	multicast.sendMessage(conn, connectionMessage)

	multicast.connections[node.identifier].outbound = true

	fmt.Println("Connection message sent, waiting for ready")

	for !multicast.ready() {
	}

	fmt.Println("Ready, sleeping")

	time.Sleep(time.Duration(5) * time.Second)

	for {
		message := <-multicast.writer
		_, err = fmt.Fprintln(conn, message)
		if err != nil {
			return
		}
		fmt.Println("SEND ", message)
	}
}

func (multicast *BasicMulticast) handleConnection(conn net.Conn) {
	fmt.Println("connection hander...")

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("err", err)
		return
	}

	fmt.Println("Line received", line)

	message := multicast.parseMessage(line)

	var node *MPNode

	for _, otherNode := range multicast.otherNodes {
		if otherNode.identifier == message.node {
			node = otherNode
			break
		}
	}

	connection := multicast.connections[node.identifier]
	connection.inbound = true

	if !connection.outbound {
		go multicast.connect(node)
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		// message := multicast.parseMessage(line)

		fmt.Print("RECV ", line)
	}
}
