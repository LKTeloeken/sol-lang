.PHONY: build test run tac clean

build:
	go build -o sollang ./src/main.go

test:
	go test ./...

run: build
	./sollang examples/conta_bancaria.sol

tac: build
	./sollang -tac examples/conta_bancaria.sol

clean:
	rm -f sollang output.tac
