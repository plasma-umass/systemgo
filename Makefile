COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html

all: build test
build: deps generate init systemctl

deps:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/coreos/go-systemd/unit
	go get -d ./...

coverdeps: deps
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/wadey/gocovmerge

generate:
	go generate ./...
init: *.go system/*.go unit/*.go lib/test/*.go lib/errors/*.go
	go build -o bin/init
systemctl: systemctl/cmd/*.go systemctl/main.go lib/systemctl/*.go
	go build -o bin/systemctl ./systemctl
install:
	go get -v ./...

test: generate
	go test ./...
cover: coverdeps generate
	./coverage.sh --html
coveralls: coverdeps generate 
	./coverage.sh --coveralls
travis: coverdeps build
	./coverage.sh --coveralls travis-ci

clean: cleancover cleanbin cleanstringers

cleanbin:
	-rm -rf bin
cleanstringers:
	-rm `find -name '*_string.go'`
cleancover:
	-rm -f $(COVER_PROFILE)

.PHONY: all generate test testdeps deps cover systemctl init install clean cleanbin cleanstringers cleancover 
