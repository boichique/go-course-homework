build:
	go build -o ./bin/client ./client && go build -o ./bin/server ./server
run-server:
	RAND_SEED=42 ./bin/server -port 8080 -min "-100" -max "100"
run-client:
	./bin/client -port 8080 -host "localhost"

# Kill server. Suppress errors and always succeed
kill-server:
	killall ./bin/server 2>/dev/null || true

# Full cycle play. Kill current server if exists and build both. Run server in background, then run client and kill server eventually
play: kill-server build
	(make run-server &) && make run-client; make kill-server
