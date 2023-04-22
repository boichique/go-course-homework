build:
	go build -o ./bin/watchcmd .

run: build
	./bin/watchcmd -interval 1s date

kill:
	kill -s SIGINT $(pgrep watchcmd)