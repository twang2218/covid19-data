.PHONY:	build run run-no-cache test clean

build:
	go build -o shanghai-crawler ./cmd

run:
	go run ./cmd -v daily

run-no-cache:
	go run ./cmd -v daily --no-cache

test:
	go test -v ./...

clean:
	go clean
	rm -f ./shanghai-crawler
	rm -rf ../data/{.web_cache}

