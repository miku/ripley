ripley: cmd/ripley/main.go
	go get ./...
	go get golang.org/x/tools/cmd/goimports
	goimports -w .
	go build -o ripley cmd/ripley/main.go

clean:
	rm -f ripley
