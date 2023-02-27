package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var balancesStrings = ""

func SetupCloseHandler(identifier string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		f, err := os.Create(identifier + "_log.txt")
		if err != nil {
			return
		}

		defer f.Close()

		f.WriteString(balancesStrings)
		os.Exit(0)
	}()
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

	genMessages := make(chan ReliableMessage)
	go transactionsToMessages(currentNode, genTransactions, genMessages)

	multicast := &ISISMulticast{currentNode: currentNode, otherNodes: otherNodes, writer: genMessages}
	err = multicast.Setup()
	if err != nil {
		return err
	}

	balances := make(map[string]int)
	SetupCloseHandler(identifier)
	for {
		message := <-multicast.Receiver()
		// fmt.Println("DELIVERED: ", message.Identifier)

		transaction := message.Transaction

		if _, contains := balances[transaction.DestAccount]; !contains {
			balances[transaction.DestAccount] = 0
		}

		if transaction.Deposit {
			balances[transaction.DestAccount] += transaction.Amount
		} else {
			if _, contains := balances[transaction.SourceAccount]; contains {
				if balances[transaction.SourceAccount] >= transaction.Amount {
					balances[transaction.SourceAccount] -= transaction.Amount
					balances[transaction.DestAccount] += transaction.Amount
				}
			}
		}

		balanceString := balancesToString(balances)
		fmt.Print(balanceString)
		balancesStrings += balanceString
	}
}

func transactionsToMessages(node *MPNode, transactions chan Transaction, messages chan ReliableMessage) {
	for {
		transaction := <-transactions
		fmt.Println("picked up transaction", transaction)
		message := ReliableMessage{Node: node.identifier, Transaction: transaction, Identifier: ""}
		// fmt.Println("ISIS SENDING ", node.Identifier)
		fmt.Println("converted to message", message)
		messages <- message
	}
}

func balancesToString(balances map[string]int) string {
	accounts := make([]string, len(balances))

	i := 0
	for k := range balances {
		accounts[i] = k
		i++
	}

	sort.Strings(accounts)
	out := "BALANCES "
	for i, account := range accounts {
		if balances[account] <= 0 {
			continue
		}

		out += account + ":" + strconv.Itoa(balances[account])
		if i != len(balances)-1 {
			out += " "
		}
	}
	out += "\n"
	return out
}
