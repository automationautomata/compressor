BINARY_NAME=compressor

CMD_DIR=compressor/

BIN_DIR=bin/

.PHONY: all build clean run

all: build

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run: build
	./$(CMD_DIR)

clean:
	rm -f $(BIN_DIR)/$(BINARY_NAME)
