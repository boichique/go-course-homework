path ?= ".." # default value is the whole apps repo
mode ?= "sequential"

build:
	go build -o ./bin/wordscount .
create-profiles-dir:
	mkdir profiles 2>/dev/null || true
run: build create-profiles-dir
	./bin/wordscount -path $(path) -mode $(mode) -cpu-profile ./profiles/cpu.prof -trace ./profiles/trace.prof
cpu-profile:
	go tool pprof -http=:8080 ./bin/wordscount ./profiles/cpu.prof
trace-profile:
	go tool trace -http=:8081 ./profiles/trace.prof