PORT?=8000

clean:

	rm -rf ./bin/execution/*
	rm -rf ./bin/binance/*

build: clean

	go build -o ./bin/execution/execution ./cmd/execution/execution.go
	cp -n ./etc/execution.properties ./bin/execution/

	go build -o ./bin/binance/binance ./cmd/binance/binance.go
	cp -n ./etc/binance.properties ./bin/binance/

test:
	go test -race ./src/msq.ai/...
