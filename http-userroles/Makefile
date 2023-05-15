# Makefile constants
DB_CONTAINER_NAME = userroles
DB_PORT = 5432
DB_USER = dbuser
DB_PASSWORD = dbpass
DB_NAME = userroles
DB_HOST ?= localhost
SERVER_ADDR ?= ":8080"
SERVER_TLS_ADDR ?= ":8443"
NETWORK_NAME = "userroles"


stop-db:
	docker stop $(DB_CONTAINER_NAME) 2>/dev/null || true


run-db: stop-db

	docker run \
		--name $(DB_CONTAINER_NAME) \
		-p 10000:$(DB_PORT) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		--rm \
		-d \
		--network $(NETWORK_NAME) \
		--mount source=userroles-db-data,target=/var/lib/postgresql/data \
		postgres:15


run-server:
	cd server && go run . -listener-type tcp -addr $(SERVER_ADDR) -conn-string "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"


run-https-server:
	cd server && go run . -tls-addr $(SERVER_TLS_ADDR) -cert "./config/server.crt" -key "./config/server.key" -conn-string "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"


stop-docker-server:
	docker stop "userroles-server" 2>/dev/null || true


run-docker-server: stop-docker-server

	docker run \
		--name "userroles-server" \
		-p 10001:8080 \
		--rm \
		-d \
		--network $(NETWORK_NAME) \
		http-userroles-server:f4 \
		-conn-string "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_CONTAINER_NAME):$(DB_PORT)/$(DB_NAME)" \
		-listener-type "tcp" \
		-addr ":8080"


run-cli:
	docker run \
		--name "userroles-cli" \
		--rm \
		-d \
		--network $(NETWORK_NAME) \
		http-userroles-cli:f4 sleep 600
