DIRS = . 

ci:
	go build $(DIRS)
	go test $(DIRS)

build:
	go build $(DIRS)

test:
	go test $(DIRS)

vet: 
	go vet $(DIRS)

fmt:
	go fmt $(DIRS)