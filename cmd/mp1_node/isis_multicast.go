package main

import (
	"bytes"
	"container/heap"
	"encoding/gob"
	"fmt"
	"strings"
	"time"
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
	lifeTime      int64
}

type Content struct {
	Proposal   bool
	Agreed     bool
	Priority   ISISPriority
	Identifier string
}

type ISISMulticast struct {
	reliableMulticast  *ReliableMulticast
	currentNode        *MPNode
	otherNodes         []*MPNode
	writer             chan ReliableMessage
	receiver           chan ISISMessage
	reliableReceiver   chan ReliableMessage
	queue              PriorityQueue
	priorities         map[string]ISISPriority
	proposals          map[string][]string
	highestPriorityNum int
	allNodes           []*MPNode
}

func (multicast *ISISMulticast) Setup() error {
	multicast.receiver = make(chan ISISMessage)
	multicast.queue = make(PriorityQueue, 0)
	multicast.priorities = make(map[string]ISISPriority)
	multicast.proposals = make(map[string][]string)

	multicast.reliableMulticast = &ReliableMulticast{currentNode: multicast.currentNode, otherNodes: multicast.otherNodes, writer: multicast.writer}
	err := multicast.reliableMulticast.Setup()
	if err != nil {
		return err
	}

	multicast.reliableReceiver = multicast.reliableMulticast.receiver
	multicast.allNodes = multicast.reliableMulticast.basicMulticast.allNodes

	heap.Init(&multicast.queue)
	go multicast.addISISReceives()

	return nil
}

func (multicast *ISISMulticast) Receiver() <-chan ISISMessage {
	return multicast.receiver
}

func Contains(slice []string, elem string) bool {
	for _, cur := range slice {
		if cur == elem {
			return true
		}
	}
	return false
}

func (multicast *ISISMulticast) addISISReceives() {

	for {
		// fmt.Println("<-I-multicast.reliableReceiver")

		message := <-multicast.reliableReceiver

		isis := ReliableToISIS(message)
		// fmt.Println("reliable to isis: ", isis)

		if !isis.proposal && !isis.agreed {
			// fmt.Println("not a proposal and not agreed (multicasted message)", isis)
			multicast.highestPriorityNum = multicast.highestPriorityNum + 1
			proposal := ISISMessage{
				Node:          multicast.currentNode.identifier,
				ToNode:        &isis.Node,
				Transaction:   isis.Transaction,
				proposal:      true,
				agreed:        false,
				priority:      ISISPriority{Num: multicast.highestPriorityNum, Identifier: multicast.currentNode.identifier},
				Identifier:    isis.Identifier,
				undeliverable: true,
				lifeTime:      time.Now().Unix(),
			}

			fmt.Println("before send proposal to multicasted writer", proposal)
			multicast.writer <- ISISToReliable(proposal)
			// fmt.Println("after send proposal to multicasted writer", proposal)

			// fmt.Println("before pushing ", isis, " to PQ", multicast.queue.Print())
			// multicast.queue.Push(&isis)
			heap.Push(&multicast.queue, &proposal)
			// fmt.Println("after pushing ", isis, " to PQ", multicast.queue.Print())
		} else if isis.proposal {
			if _, contains := multicast.priorities[isis.Transaction.Identifier]; !contains {
				multicast.priorities[isis.Transaction.Identifier] = isis.priority
			} else if ComparePriorities(isis.priority, multicast.priorities[isis.Transaction.Identifier]) {
				multicast.priorities[isis.Transaction.Identifier] = isis.priority
			}

			if _, contains := multicast.proposals[isis.Transaction.Identifier]; !contains {
				multicast.proposals[isis.Transaction.Identifier] = make([]string, 0)
			}
			multicast.proposals[isis.Transaction.Identifier] = append(multicast.proposals[isis.Transaction.Identifier], isis.Node)

			// fmt.Println("RECEIVED PROPOSAL: " + isis.Transaction.Identifier + " (" + strconv.Itoa(multicast.proposedCounts[isis.Transaction.Identifier]) + ")")
			// fmt.Println("recieved proposal cnt: ", multicast.proposedCounts[isis.Transaction.identifier])
			// if all nodes that havent failed, have sent a proposal

			fmt.Println("proposals: ", multicast.proposals[isis.Transaction.Identifier])
			receivedAllProposals := true
			for _, node := range multicast.allNodes {

				fmt.Println("Failed nodes: ", multicast.reliableMulticast.basicMulticast.failedNodes)
				multicast.reliableMulticast.basicMulticast.failedNodesLock.Lock()
				if _, contains := multicast.reliableMulticast.basicMulticast.failedNodes[node.identifier]; contains {
					multicast.reliableMulticast.basicMulticast.failedNodesLock.Unlock()
					fmt.Println("Failed node being skipped: ", node.identifier)
					continue
				}
				multicast.reliableMulticast.basicMulticast.failedNodesLock.Unlock()

				if !Contains(multicast.proposals[isis.Transaction.Identifier], node.identifier) {
					receivedAllProposals = false
					break
				}
			}

			if receivedAllProposals {
				// fmt.Println("SENDING AGREE: " + isis.Transaction.Identifier + " [" + strconv.Itoa(multicast.priorities[isis.Transaction.Identifier].Num) + "," + multicast.priorities[isis.Transaction.Identifier].Identifier + "]")
				agreed := ISISMessage{
					Node:          multicast.currentNode.identifier,
					Transaction:   isis.Transaction,
					priority:      multicast.priorities[isis.Transaction.Identifier],
					agreed:        true,
					proposal:      false,
					undeliverable: false, //not useuful
					Identifier:    isis.Identifier,
				}
				multicast.writer <- ISISToReliable(agreed)
			}
		} else if isis.agreed {
			fmt.Println("agreed on", isis)
			if isis.priority.Num > multicast.highestPriorityNum {
				multicast.highestPriorityNum = isis.priority.Num
			}

			// fmt.Println("before update", isis, multicast.queue.Print())
			multicast.queue.update(&isis, isis.priority)
			fmt.Println("after update", isis, multicast.queue.Print())

			// deliver all deliverable at front of queue
			for len(multicast.queue) > 0 {
				message := heap.Pop(&multicast.queue).(*ISISMessage)
				if !message.undeliverable {
					multicast.receiver <- *message
				} else {
					multicast.reliableMulticast.basicMulticast.failedNodesLock.Lock()
					if _, contains := multicast.reliableMulticast.basicMulticast.failedNodes[message.Node]; contains {
						multicast.reliableMulticast.basicMulticast.failedNodesLock.Unlock()
						continue
					} else {
						multicast.reliableMulticast.basicMulticast.failedNodesLock.Unlock()
					}

					if time.Now().Unix()-message.lifeTime > 10 {
						continue
					}

					heap.Push(&multicast.queue, message)
					break
				}
			}
		}
	}
}

//aj thoughts
// if agreed == false and proposal == false, propose a priority and multicast it
// if agreed == true and proposal == false, parse agreed priority, store in priority queue as deliverable?
// if agreed == false and proposal == true, parse proposal,

func ReliableToISIS(message ReliableMessage) ISISMessage {
	decoder := gob.NewDecoder(strings.NewReader(message.Content))

	content := &Content{}
	err := decoder.Decode(content)
	// fmt.Println("recieved cotentn: ", message.Content)
	if err != nil {
		return ISISMessage{
			proposal:      false,
			agreed:        false,
			Node:          message.Node,
			Identifier:    message.Identifier,
			Transaction:   message.Transaction,
			ToNode:        message.ToNode,
			undeliverable: true,
		}
	}

	// fmt.Println("actual isis messaged received")

	return ISISMessage{
		proposal:      content.Proposal,
		agreed:        content.Agreed,
		priority:      content.Priority,
		undeliverable: true, //important bc all messages added to queue should initially be undeliverable
		Node:          message.Node,
		Identifier:    message.Identifier,
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
		Identifier:  message.Identifier,
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
