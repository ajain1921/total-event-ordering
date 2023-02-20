package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type Node struct {
	identifier string
	hostname   string
	port       string
	connected  bool
}

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

	var currentNode *Node
	otherNodes := []*Node{}
	err := parseConfig(fileName, identifier, &currentNode, &otherNodes)
	if err != nil {
		return err
	}
	// fmt.Println(otherNodes)

	ln, err := net.Listen("tcp", ":"+currentNode.port)
	if err != nil {
		return err
	}

	for _, node := range otherNodes {
		fmt.Println(node)
		go connect(node)

		go handleConnection(node, ln)
	}
	for {

	}
	return nil
}

func parseConfig(filename string, identifier string, currentNode **Node, otherNodes *[]*Node) error {
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

		newNode := &Node{identifier: id, hostname: hostname, port: port, connected: false}
		if id == identifier {
			*currentNode = newNode
			continue
		}
		*otherNodes = append(*otherNodes, newNode)
	}

	// fmt.Println(otherNodes)

	readFile.Close()
	return nil
}

func connect(node *Node) {
	conn, err := net.Dial("tcp", node.hostname+":"+node.port)
	if err != nil {
		return
	}
	defer conn.Close()

	fmt.Fprintln(conn, "YO I CONNEcted to "+node.identifier)

	// reader := bufio.NewReader(conn)
}

func handleConnection(node *Node, ln net.Listener) {
	conn, err := ln.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ruh roh: %v\n", err)
	}

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// TODO: Handle failure
			return
		}
		fmt.Println(line)
	}
}
