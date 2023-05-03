# Makefile constants
DB_CONTAINER_NAME = northwinddb
DB_PORT = 5432
DB_USER = nwuser
DB_PASSWORD = nwpass
DB_NAME = northwind

stop-db:
	# Stop the docker container if it is running
		# 2>/dev/null: redirect stderr to /dev/null to avoid printing error message if the container is not running
		# || true: make sure the command exits with 0 status code
	docker stop $(DB_CONTAINER_NAME) 2>/dev/null || true

# Execute the stop-db target before running the run-db target
run-db: stop-db
	# Run the docker container
		# --name: name of the container
		# -p: map the port of the container to the port of the host
		# -e: set env variable in the container to specify the admin user, password, and the database name to create on start
		# --rm: remove the container when it exits
		# -d: run the container in the background
		# postgres:15: the image to run. This is the latest version of the Official Postgres image
	docker run \
		--name $(DB_CONTAINER_NAME) \
		-p $(DB_PORT):$(DB_PORT) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		--rm \
		-d \
		postgres:15

# ?= is used to set a default value for a variable if it is not already set. So here the default value is false
INIT ?= false
run-queries:
	go run . -db-host localhost -db-port $(DB_PORT) -db-user $(DB_USER) -db-password $(DB_PASSWORD) -db-name $(DB_NAME) -init $(INIT)

.PHONY: stop-db, run-db, run-queries
