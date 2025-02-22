.PHONY: test

test:
	go test -v ./...

.PHONY: simtest
simtest:
	GOEXPERIMENT=synctest go test -v ./... -tags=simtest
