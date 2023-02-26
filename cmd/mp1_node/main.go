package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	if len(args) != 2 {
		return errors.New("exactly 2 command line argument please")
	}

	identifier := args[0]
	fileName := args[1]

	var currentNode *MPNode
	otherNodes := []*MPNode{}
	err := ParseConfig(fileName, identifier, &currentNode, &otherNodes)
	if err != nil {
		return err
	}

	genTransactions := make(chan Transaction)
	go StreamTransactions(os.Stdin, genTransactions)

	genMessages := make(chan Message)
	go transactionsToMessages(currentNode, genTransactions, genMessages)

	multicast := &BasicMulticast{currentNode: currentNode, otherNodes: otherNodes, writer: genMessages}
	err = multicast.Setup()
	if err != nil {
		return err
	}

	for {
		message := <-multicast.Receiver()
		// TODO: Apply transaction
		fmt.Println(message)
	}
}

func transactionsToMessages(node *MPNode, transactions chan Transaction, messages chan Message) {
	for {
		<-transactions
		messages <- Message{node: node.identifier, content: "gobby gob"}
	}
}
