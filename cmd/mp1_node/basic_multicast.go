package main

import (
	"encoding/gob"
	"net"
	"sync"
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

type BasicMessage struct {
	Node        string
	Content     string
	Transaction Transaction
	ToNode      *string
}

type BasicMulticast struct {
	currentNode     *MPNode
	otherNodes      []*MPNode
	allNodes        []*MPNode
	connections     map[string]*ConnectionStatus
	writer          chan BasicMessage
	receiver        chan BasicMessage
	channels        map[string]chan BasicMessage
	failedNodes     map[string]interface{}
	failedNodesLock sync.Mutex
	connectionsLock sync.Mutex
}

func (multicast *BasicMulticast) Setup() error {
	multicast.receiver = make(chan BasicMessage)
	multicast.connections = make(map[string]*ConnectionStatus)
	multicast.channels = make(map[string]chan BasicMessage)
	multicast.failedNodes = make(map[string]interface{})

	ln, err := net.Listen("tcp", multicast.currentNode.hostname+":"+multicast.currentNode.port)
	if err != nil {
		return err
	}

	var allNodes []*MPNode
	allNodes = append(allNodes, multicast.otherNodes...)
	allNodes = append(allNodes, multicast.currentNode)

	multicast.allNodes = allNodes

	// allNodes = append(allNodes, multicast.currentNode...)

	for _, node := range allNodes {
		multicast.channels[node.identifier] = make(chan BasicMessage)
		multicast.connections[node.identifier] = &ConnectionStatus{}
	}

	for _, node := range allNodes {
		go multicast.connect(node, multicast.channels[node.identifier])
	}

	go multicast.forward()

	for range allNodes {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go multicast.handleConnection(conn)
	}

	return nil
}

func (multicast *BasicMulticast) Receiver() <-chan BasicMessage {
	return multicast.receiver
}

func (multicast *BasicMulticast) ready() bool {
	multicast.connectionsLock.Lock()
	for _, node := range multicast.allNodes {
		connectionStatus := multicast.connections[node.identifier]
		if !connectionStatus.inbound || !connectionStatus.outbound {
			multicast.connectionsLock.Unlock()
			return false
		}
	}
	multicast.connectionsLock.Unlock()
	return true
}

func (multicast *BasicMulticast) forward() {
	for {
		message := <-multicast.writer
		for _, channel := range multicast.channels {

			channel <- message
		}
	}
}

func (multicast *BasicMulticast) connect(node *MPNode, channel chan BasicMessage) {
	// connection := multicast.connections[node.identifier]

	// if connection.outbound {
	// 	return
	// }

	conn, err := net.Dial("tcp", node.hostname+":"+node.port)
	if err != nil {
		// fmt.Println("trying to connect to " + node.identifier + " but failed")
		return
	}
	defer conn.Close()

	multicast.connectionsLock.Lock()

	if multicast.connections[node.identifier].outbound {
		multicast.connectionsLock.Unlock()
		return
	}
	multicast.connections[node.identifier].outbound = true

	multicast.connectionsLock.Unlock()

	encoder := gob.NewEncoder(conn)

	connectionMessage := BasicMessage{Node: multicast.currentNode.identifier, Content: "connected"}
	err = encoder.Encode(connectionMessage)
	if err != nil {
		// fmt.Println(err)
		return
	}
	// fmt.Println("SEND ", connectionMessage)

	// fmt.Println("Connection message sent, waiting for ready")

	for !multicast.ready() {
	}

	// fmt.Println("Ready, sleeping")

	time.Sleep(time.Duration(5) * time.Second)

	wg.Done()

	for {
		message := <-channel
		// If we're unicasting and this isn't the destination... STOP
		if message.ToNode != nil && *message.ToNode != node.identifier {
			continue
		}
		err = encoder.Encode(&message)
		if err != nil {
			// fmt.Println(err)
			break
		}
		// fmt.Println("SEND ", message)
	}

	multicast.failedNodesLock.Lock()
	multicast.failedNodes[node.identifier] = struct{}{}
	multicast.failedNodesLock.Unlock()

	for {
		<-channel
	}
}

func (multicast *BasicMulticast) handleConnection(conn net.Conn) {
	// fmt.Println("connection hander...")

	dec := gob.NewDecoder(conn)

	message := &BasicMessage{}
	err := dec.Decode(message)
	if err != nil {
		// fmt.Println(err)
		return
	}
	// fmt.Println("RECV ", *message)
	var node *MPNode

	for _, otherNode := range multicast.allNodes {
		if otherNode.identifier == message.Node {
			node = otherNode
			break
		}
	}

	multicast.connectionsLock.Lock()
	connection := multicast.connections[node.identifier]
	connection.inbound = true

	if !connection.outbound {
		multicast.connectionsLock.Unlock()
		go multicast.connect(node, multicast.channels[node.identifier])
	} else {
		multicast.connectionsLock.Unlock()
	}

	for {
		message := &BasicMessage{}
		err = dec.Decode(message)
		if err != nil {
			// fmt.Println(err)
			return
		}

		multicast.receiver <- *message

		// fmt.Println("RECV ", *message)
	}
}
