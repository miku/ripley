ripley: cmd/ripley/main.go
	goimports -w .
	go build -o ripley cmd/ripley/main.go

clean:
	rm -f ripley

	
