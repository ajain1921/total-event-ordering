package main

import (
	"strconv"
	"strings"
)

type ReliableMessage struct {
	Node        string
	Content     string
	Identifier  string
	Transaction Transaction
	ToNode      *string
}

type ReliableMulticast struct {
	basicMulticast *BasicMulticast
	currentNode    *MPNode
	otherNodes     []*MPNode
	writer         chan ReliableMessage
	receiver       chan ReliableMessage
	basicWriter    chan BasicMessage
	basicReceiver  chan BasicMessage
	received       map[string]interface{}
}

var counter = 0

func (multicast *ReliableMulticast) Setup() error {
	multicast.received = make(map[string]interface{})
	multicast.receiver = make(chan ReliableMessage)
	multicast.basicWriter = make(chan BasicMessage)

	multicast.basicMulticast = &BasicMulticast{currentNode: multicast.currentNode, otherNodes: multicast.otherNodes, writer: multicast.basicWriter}
	err := multicast.basicMulticast.Setup()
	if err != nil {
		return err
	}

	multicast.basicReceiver = multicast.basicMulticast.receiver

	go multicast.addReliableReceives()
	go multicast.addReliableWrites()

	return nil
}

func (multicast *ReliableMulticast) Receiver() <-chan ReliableMessage {
	return multicast.receiver
}

func (multicast *ReliableMulticast) addReliableReceives() {
	for {
		message := <-multicast.basicReceiver

		reliable := basicToReliable(message)

		if _, contains := multicast.received[reliable.Identifier]; !contains {
			multicast.received[reliable.Identifier] = true
			if message.Node != multicast.currentNode.identifier {
				multicast.basicWriter <- message
			}
			multicast.receiver <- reliable
		}
	}
}

func (multicast *ReliableMulticast) addReliableWrites() {
	for {
		message := <-multicast.writer

		message.Identifier = multicast.currentNode.identifier + "," + strconv.Itoa(counter)
		counter++

		basic := reliableToBasic(message)
		multicast.basicWriter <- basic
		// Send to self!
		multicast.basicReceiver <- basic
	}
}

func basicToReliable(basic BasicMessage) ReliableMessage {
	split := strings.SplitN(basic.Content, ":", 2)

	// fmt.Println("received large content: ", split[1], " BREAK ", basic.Content)

	return ReliableMessage{
		Node:        basic.Node,
		Content:     split[1],
		Identifier:  split[0],
		Transaction: basic.Transaction,
		ToNode:      basic.ToNode,
	}
}

func reliableToBasic(reliable ReliableMessage) BasicMessage {
	content := reliable.Identifier + ":" + reliable.Content

	// fmt.Println("Sending legit: " + reliable.Content)

	return BasicMessage{
		Node:        reliable.Node,
		Content:     content,
		Transaction: reliable.Transaction,
		ToNode:      reliable.ToNode,
	}
}
