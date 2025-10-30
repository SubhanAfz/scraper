run:
	go run cmd/api/main.go
build: clean
	go build -o bin/api cmd/api/main.go
run-mcp:
	go run cmd/mcp/main.go
build-mcp: clean
	go build -o bin/mcp cmd/mcp/main.go
clean:
	rm -rf bin
	mkdir bin
	cp rules.json bin/rules.json