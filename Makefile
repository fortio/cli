
wait-demo:
	# Doesn't work with tinygo (yet) because undefined symbol: _os/signal.signal_enable...
	go run -tags no_net,no_json ./sampleTool/ -wait arg1 arg2

lint: .golangci.yml
	golangci-lint run

.golangci.yml: Makefile
	curl -fsS -o .golangci.yml https://raw.githubusercontent.com/fortio/workflows/main/golangci.yml

.PHONY: lint wait-demo
