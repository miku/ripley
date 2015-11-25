ripley: cmd/ripley/main.go
	go get ./...
	goimports -w .
	go build -o ripley cmd/ripley/main.go

clean:
	rm -f ripley
