install:
	GO111MODULE=on go install

test:
	GO111MODULE=on go test -v ...

PHONY: install
