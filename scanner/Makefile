GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

scanner:
	go build -o $(GOBIN)/scanner.exe ./cmd/scanner/ 
	@echo "Done building."
	@echo "Run \"$(GOBIN)/crawler\" to launch scanner."
	
clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

.PHONY: scanner clean