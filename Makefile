export GOPATH = $(CURDIR)/vendor:$(CURDIR)


all: bin/bakapy-scheduler bin/bakapy-show-meta bin/bakapy-run-job

bin/bakapy-scheduler:
	go install bakapy/cmd/bakapy-scheduler

bin/bakapy-show-meta:
	go install bakapy/cmd/bakapy-show-meta

bin/bakapy-run-job:
	go install bakapy/cmd/bakapy-run-job

test:
	go test -covermode=count -coverprofile=coverage.out --run=. bakapy

racetest:
	go test -race -covermode=count -coverprofile=coverage.out --run=. bakapy

clean:
	rm -rf bin/ pkg/ vendor/bin vendor/pkg

.PHONY: bin/bakapy-scheduler bin/bakapy-show-meta bin/bakapy-run-job test racetest clean
