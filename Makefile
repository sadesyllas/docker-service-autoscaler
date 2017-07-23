BIN_FILENAME := docker-service-autoscaler

all: clean linux win

linux: clean_linux
	mkdir -p ./bin/linux
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/$(BIN_FILENAME) ./src/main

clean_linux:
	rm -rf bin/linux

win: clean_win
	mkdir -p ./bin/win
	GOOS=windows GOARCH=amd64 go build -o ./bin/win/$(BIN_FILENAME).exe ./src/main

clean_win:
	rm -rf bin/win

clean:
	rm -rf bin

