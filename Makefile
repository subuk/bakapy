GO=go
export GOPATH = $(CURDIR)/vendor:$(CURDIR)


all: bin/bakapy-scheduler bin/bakapy-show-meta bin/bakapy-run-job

bin/bakapy-scheduler:
	$(GO) install bakapy/cmd/bakapy-scheduler

bin/bakapy-show-meta:
	$(GO) install bakapy/cmd/bakapy-show-meta

bin/bakapy-run-job:
	$(GO) install bakapy/cmd/bakapy-run-job

test:
	$(GO) test -covermode=count -coverprofile=coverage.out --run=. bakapy

racetest:
	$(GO) test -race -covermode=count -coverprofile=coverage.out --run=. bakapy

clean:
	rm -rf bin/ pkg/ vendor/bin vendor/pkg

package-%: Dockerfile.%
	docker build -f "Dockerfile.$*" -t "bakapy-build-$*" .
	rm -rf "native-packages/$*"
	mkdir -p "native-packages/$*"
	docker run --rm "bakapy-build-$*" /bin/bash -c 'tar -C /packages -cf - .' | tar -C "./native-packages/$*" -xf -

package-all: package-trusty package-precise package-wheezy package-centos6

.PHONY: bin/bakapy-scheduler bin/bakapy-show-meta bin/bakapy-run-job test racetest clean package-all package-%
