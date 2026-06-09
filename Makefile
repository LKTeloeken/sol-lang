.PHONY: build test run clean

build:
	go build -o sollang ./src/main.go

test:
	go test ./...

clean:
	rm -f sollang
