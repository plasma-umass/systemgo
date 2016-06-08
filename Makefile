all:
	go generate ./...
	go build -o bin/init
	go build -o bin/systemctl ./systemctl

.PHONY: clean
clean:
	rm `find -name '*_string.go'`
	rm -rf bin

.PHONY: test
test:
	go test ./...
