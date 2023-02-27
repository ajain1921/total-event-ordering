package main

import (
	"bytes"
	"encoding/gob"
	"strings"
)

type ISISPriority struct {
	Num        int
	Identifier string
}

type ISISMessage struct {
	// Common with ReliableMessage
	Node        string
	Identifier  string
	Transaction Transaction
	ToNode      *string

	// Content
	proposal      bool
	agreed        bool
	priority      ISISPriority
	undeliverable bool // only messages that are deliverable can be popped off queue
	index         int
}

type Content struct {
	Proposal   bool
	Agreed     bool
	Priority   ISISPriority
	Identifier string
}

type ISISMulticast struct {
	reliableMulticast *ReliableMulticast
	currentNode       *MPNode
	otherNodes        []*MPNode
	writer            chan ReliableMessage
	receiver          chan ISISMessage
	reliableReceiver  chan ReliableMessage
	queue             PriorityQueue
	priorities        map[string]ISISPriority
	proposedCounts    map[string]int
	highestPriority   ISISPriority
}

func (multicast *ISISMulticast) Setup() error {
	multicast.receiver = make(chan ISISMessage)
	multicast.queue = make(PriorityQueue, 0)
	multicast.priorities = make(map[string]ISISPriority)
	multicast.proposedCounts = make(map[string]int)

	multicast.reliableMulticast = &ReliableMulticast{currentNode: multicast.currentNode, otherNodes: multicast.otherNodes, writer: multicast.writer}
	err := multicast.reliableMulticast.Setup()
	if err != nil {
		return err
	}

	multicast.reliableReceiver = multicast.reliableMulticast.receiver

	go multicast.addISISReceives()

	return nil
}

func (multicast *ISISMulticast) Receiver() <-chan ISISMessage {
	return multicast.receiver
}

func (multicast *ISISMulticast) addISISReceives() {

	for {
		message := <-multicast.reliableReceiver

		isis := reliableToISIS(message)

		if !isis.proposal && !isis.agreed {
			multicast.highestPriority.Num++
			// fmt.Println("sending proposal")
			proposal := ISISMessage{
				Node:        multicast.currentNode.identifier,
				ToNode:      &isis.Node,
				Transaction: isis.Transaction,
				proposal:    true,
				agreed:      false,
				priority:    multicast.highestPriority,
				Identifier:  isis.Identifier,
			}
			multicast.writer <- ISISToReliable(proposal)
			multicast.queue.Push(&isis)
		} else if isis.proposal {
			if _, contains := multicast.priorities[isis.Identifier]; !contains {
				multicast.priorities[isis.Identifier] = isis.priority
			} else if ComparePriorities(isis.priority, multicast.priorities[isis.Identifier]) {
				multicast.priorities[isis.Identifier] = isis.priority
			}
			multicast.proposedCounts[isis.Identifier]++
			// fmt.Println("recieved proposal cnt: ", multicast.proposedCounts[isis.Identifier])
			if multicast.proposedCounts[isis.Identifier] >= len(multicast.otherNodes)+1 {
				// fmt.Println("Sending agree")
				agreed := ISISMessage{
					Node:          multicast.currentNode.identifier,
					Transaction:   isis.Transaction,
					priority:      multicast.priorities[isis.Identifier],
					agreed:        true,
					proposal:      false,
					undeliverable: false, //not useuful
					Identifier:    isis.Identifier,
				}
				multicast.writer <- ISISToReliable(agreed)
			}
		} else if isis.agreed {
			if ComparePriorities(isis.priority, multicast.highestPriority) {
				multicast.highestPriority = isis.priority
			}
			multicast.queue.update(&isis, isis.priority)
			isis.undeliverable = false
			// deliver all deliverable at front of queue
			for !multicast.queue.Peek().(*ISISMessage).undeliverable {
				message := multicast.queue.Pop().(*ISISMessage)
				multicast.receiver <- *message
			}
		}
	}
}

//aj thoughts
// if agreed == false and proposal == false, propose a priority and multicast it
// if agreed == true and proposal == false, parse agreed priority, store in priority queue as deliverable?
// if agreed == false and proposal == true, parse proposal,

func reliableToISIS(message ReliableMessage) ISISMessage {
	decoder := gob.NewDecoder(strings.NewReader(message.Content))

	content := &Content{}
	err := decoder.Decode(content)
	// fmt.Println("recieved cotentn: ", message.Content)
	if err != nil {
		return ISISMessage{
			proposal:    false,
			agreed:      false,
			Node:        message.Node,
			Identifier:  message.Identifier,
			Transaction: message.Transaction,
			ToNode:      message.ToNode,
		}
	}

	// fmt.Println("actual isis messaged received")

	return ISISMessage{
		proposal:      content.Proposal,
		agreed:        content.Agreed,
		priority:      content.Priority,
		undeliverable: false, //unimportant
		Node:          message.Node,
		Identifier:    content.Identifier,
		Transaction:   message.Transaction,
		ToNode:        message.ToNode,
	}
}

func ISISToReliable(message ISISMessage) ReliableMessage {
	// TODO: Parse reliable message into PROPER ISISMessage

	// Content format: <agreed>:<proposal>

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	contentStruct := Content{
		Proposal:   message.proposal,
		Agreed:     message.agreed,
		Priority:   message.priority,
		Identifier: message.Identifier,
	}
	err := encoder.Encode(contentStruct)
	if err != nil {
		panic("AHHHHAAAAA")
	}

	content := buffer.String()

	return ReliableMessage{
		Node:        message.Node,
		Content:     content,
		Transaction: message.Transaction,
		Identifier:  "",
		ToNode:      message.ToNode,
	}
}

// True if a > b
func ComparePriorities(a, b ISISPriority) bool {
	first := a.Num - b.Num
	if first != 0 {
		return first > 0
	}

	return strings.Compare(a.Identifier, b.Identifier) > 0
}
