package main

import (
	"encoding/gob"
	"strings"
)

type ISISPriority struct {
	num        int
	identifier string
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
	proposal bool
	agreed   bool
	priority ISISPriority
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
			multicast.highestPriority.num++
			proposal := ISISMessage{
				Node:        multicast.currentNode.identifier,
				ToNode:      &isis.Node,
				Transaction: Transaction{},
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
			if multicast.proposedCounts[isis.Identifier] >= len(multicast.otherNodes)+1 {
				agreed := ISISMessage{
					Node:          multicast.currentNode.identifier,
					Transaction:   Transaction{},
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
	if err != nil {
		panic("AHHHHHHHHH")
	}

	return ISISMessage{
		proposal:      content.proposal,
		agreed:        content.agreed,
		priority:      content.priority,
		undeliverable: false, //unimportant
		Node:          message.Node,
		Identifier:    message.Identifier,
		Transaction:   message.Transaction,
		ToNode:        message.ToNode,
	}
}

func ISISToReliable(message ISISMessage) ReliableMessage {
	// TODO: Parse reliable message into PROPER ISISMessage

	// Content format: <agreed>:<proposal>

	builder := strings.Builder{}
	encoder := gob.NewEncoder(&builder)

	err := encoder.Encode(message)
	if err != nil {
		panic("AHHHHAAAAA")
	}

	content := builder.String()

	return ReliableMessage{
		Node:        message.Node,
		Content:     content,
		Transaction: message.Transaction,
		Identifier:  message.Identifier,
		ToNode:      message.ToNode,
	}

}

// True if a > b
func ComparePriorities(a, b ISISPriority) bool {
	first := a.num - b.num
	if first != 0 {
		return first > 0
	}

	return strings.Compare(a.identifier, b.identifier) > 0
}
