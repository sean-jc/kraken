.PHONY: default up
default:
	go install -v $(DEBUG) ./...
	@echo "Release the Kraken!"

up:
	go get -t -v -d -u ./...
	@echo "Kraken has new abilities!"
