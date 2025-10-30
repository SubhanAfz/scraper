run:
	go run cmd/api/main.go
build:
	go build -o bin/api cmd/api/main.go
clean:
	rm -rf bin/api
run-mcp:
	go run cmd/mcp/main.go
build-mcp:
	go build -o bin/mcp cmd/mcp/main.go
clean-mcp:
	rm -rf bin/mcp