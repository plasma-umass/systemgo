COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html

all: build test

deps:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/coreos/go-systemd/unit
	go get -d ./...
testdeps: deps
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
coverdeps: testdeps
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/wadey/gocovmerge

build: generate init systemctl

generate: deps
	go generate ./...
init: *.go system/*.go unit/*.go lib/test/*.go lib/errors/*.go generate
	go build -o bin/init
systemctl: systemctl/cmd/*.go systemctl/main.go lib/systemctl/*.go generate
	go build -o bin/systemctl ./systemctl
install: build
	go get -v ./...

test: testdeps generate
	go test ./...
cover: coverdeps generate
	./coverage.sh --html
coveralls: coverdeps generate 
	./coverage.sh --coveralls
travis: coverdeps build
	./coverage.sh --coveralls travis-ci

clean: cleancover cleanbin cleanstringers cleanmock

cleanbin:
	-rm -rf bin
cleanstringers:
	-rm `find -name '*_string.go'`
cleancover:
	-rm -f $(COVER_PROFILE)
cleanmock:
	-rm `find -name 'mock_*_test.go'`

.PHONY: all generate test testdeps deps cover coverdeps systemctl init install clean cleanbin cleanstringers cleancover cleanmock build travis
