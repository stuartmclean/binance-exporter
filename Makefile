PROJECT_BIN = binancebot

.PHONY: build

default: build

build:
	go mod tidy && go build -v -o $(PROJECT_BIN)

