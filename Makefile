all: mp1_node
.PHONY: all

mp1_node: cmd/mp1_node/main.go
	go build -o bin/mp1_node cmd/**/*.go
