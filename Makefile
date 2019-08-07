PORT?=8000

clean:

	rm -rf ./bin/execution/*
	rm -rf ./bin/binance/*
	rm -rf ./bin/ib/*

build: clean

	go build -o ./bin/execution/execution ./cmd/execution/execution.go
	cp -n ./etc/execution.properties ./bin/execution/

	go build -o ./bin/binance/binance ./cmd/binance/binance.go
	cp -n ./etc/binance.properties ./bin/binance/

	go build -o ./bin/ib/ib ./cmd/ib/ib.go
	cp -n ./etc/ib.properties ./bin/ib/

test:
	go test -race ./src/msq.ai/...
