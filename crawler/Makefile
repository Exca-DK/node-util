GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

crawler:
	go build -o $(GOBIN)/crawler.exe ./cmd/crawler/ 
	@echo "Done building."
	@echo "Run \"$(GOBIN)/crawler\" to launch crawler."
	
clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

.PHONY: crawler clean