package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Transaction struct {
	Deposit       bool
	Amount        int
	SourceAccount string
	DestAccount   string
	Identifier    string
}

var transactionNumber = 0

func StreamTransactions(file *os.File, transactions chan Transaction, currentNode *MPNode, transactionsLog chan string) error {
	reader := bufio.NewReader(file)
	reader.Reset(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		fmt.Print("GEN: ", line)
		var transactionType string
		var sourceAccount string
		var destAccount string
		var amount int

		//try parsing as transfer
		_, err = fmt.Sscanf(line, "%s %s -> %s %d", &transactionType, &sourceAccount, &destAccount, &amount)
		if err != nil {
			//if fails, try parsing as deposit
			_, err := fmt.Sscanf(line, "%s %s %d", &transactionType, &destAccount, &amount)

			if err != nil {
				return err
			}
		}

		transaction := Transaction{
			Deposit:       transactionType == "DEPOSIT",
			Amount:        amount,
			SourceAccount: sourceAccount,
			DestAccount:   destAccount,
			Identifier:    currentNode.identifier + "_" + strconv.Itoa(transactionNumber) + "_T",
		}

		// fmt.Println("GEN: ", transaction)
		transactionsLog <- transaction.Identifier
		transactions <- transaction
		transactionNumber += 1
	}
}
