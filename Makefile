.PHONY: build test run clean

build:
	go build -o solc ./src/main.go

test:
	go test ./...

run: build
	./solc --run examples/conta_bancaria.sol

clean:
	rm -f solc output.tac output.ll program
