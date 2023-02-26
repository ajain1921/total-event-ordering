package main

import (
	"encoding/gob"
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
	Node        string
	Content     string
	Transaction Transaction
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

func (multicast *BasicMulticast) connect(node *MPNode) {
	conn, err := net.Dial("tcp", node.hostname+":"+node.port)
	if err != nil {
		fmt.Println("trying to connect to " + node.identifier + " but failed")
		return
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)

	connectionMessage := Message{Node: multicast.currentNode.identifier, Content: "connected"}
	err = encoder.Encode(connectionMessage)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("SEND ", connectionMessage)

	multicast.connections[node.identifier].outbound = true

	fmt.Println("Connection message sent, waiting for ready")

	for !multicast.ready() {
	}

	fmt.Println("Ready, sleeping")

	time.Sleep(time.Duration(5) * time.Second)

	for {
		message := <-multicast.writer
		err = encoder.Encode(&message)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("SEND ", message)
	}
}

func (multicast *BasicMulticast) handleConnection(conn net.Conn) {
	fmt.Println("connection hander...")

	dec := gob.NewDecoder(conn)

	message := &Message{}
	err := dec.Decode(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("RECV ", *message)
	var node *MPNode

	for _, otherNode := range multicast.otherNodes {
		if otherNode.identifier == message.Node {
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
		message := &Message{}
		err = dec.Decode(message)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("RECV ", *message)
	}
}
