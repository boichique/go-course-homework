PORT ?= 18080

build:
	go build -o ./bin/server ./server && go build -o ./bin/testclient ./testclient
run-server: build
	./bin/server -port $(PORT)
test-correctness: build
	./bin/testclient -mode correctness -port $(PORT)
test-concurrency: build
	./bin/testclient -mode concurrency -port $(PORT)
