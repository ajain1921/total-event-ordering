package main

import (
	"bufio"
	"fmt"
	"os"
)

type Transaction struct {
	deposit       bool
	amount        int
	sourceAccount string
	destAccount   string
}

func StreamTransactions(file *os.File, transactions chan Transaction) error {
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

		transactions <- Transaction{
			deposit:       transactionType == "DEPOSIT",
			amount:        amount,
			sourceAccount: sourceAccount,
			destAccount:   destAccount}
	}
}
