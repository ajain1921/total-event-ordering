package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ParseConfig(filename string, identifier string, currentNode **MPNode, otherNodes *[]*MPNode) error {
	readFile, err := os.Open(filename)

	if err != nil {
		return err
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	i := 0

	var id string
	var hostname string
	var port string

	for fileScanner.Scan() {
		if i == 0 {
			i += 1
			continue
		}
		line := fileScanner.Text()

		_, err = fmt.Fscanf(strings.NewReader(line), "%s %s %s", &id, &hostname, &port)
		if err != nil {
			return err
		}

		newNode := &MPNode{identifier: id, hostname: hostname, port: port}
		if id == identifier {
			*currentNode = newNode
			continue
		}
		*otherNodes = append(*otherNodes, newNode)
	}

	readFile.Close()
	return nil
}
