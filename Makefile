PORT?=8000

clean:

	rm -rf ./bin/execution/*

build: clean

	go build -o ./bin/execution/execution ./cmd/execution/execution.go

	cp -n ./etc/execution.properties ./bin/execution/
	cp -n ./etc/sql/schema.sql ./bin/execution/

test:
	go test -race ./...
