# build source and put into ./bin
build-source:
	go build -o ./bin/source ./source

# build collector and put into ./bin
build-collector:
	go build -o ./bin/collector ./collector

# build source and then run it with custom parameters and fixed seed
run-source: build-source
	RAND_SEED=1337 ./bin/source -i 200ms -d 5s -f -1000 -t 1000

# build source and collector, run source and pipe it into collector
run: build-source build-collector
	RAND_SEED=1337 ./bin/source -i 200ms -d 5s -f -1000 -t 1000 | ./bin/collector