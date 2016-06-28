COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html

all: deps generate init systemctl test

deps:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/coreos/go-systemd/unit
	go get -d ./...

testdeps: deps
	go get golang.org/x/tools/cmd/cover

generate:
	go generate ./...
init: *.go system/*.go unit/*.go lib/test/*.go lib/errors/*.go
	go build -o bin/init
systemctl: systemctl/cmd/*.go systemctl/main.go lib/systemctl/*.go
	go build -o bin/systemctl ./systemctl
install:
	go get -t ./...

test: testdeps
	go test ./...
cover:
	go test -cover ./...

# Turned out, that go test does not allow getting coverage of multiple packages
# at once - implement a script as a workaraund?
#$(COVER_PROFILE): 
#	go test -covermethod=$(COVER_METHOD) -coverprofile=$(COVER_PROFILE) ./...
#coverfunc: $(COVER_PROFILE)
#	go tool cover -func=$(COVER_PROFILE)
#coverhtml: $(COVER_PROFILE)
#	go tool cover -html=$(COVER_PROFILE)


clean: cleancover cleanbin cleanstringers

cleanbin:
	-rm -rf bin
cleanstringers:
	-rm `find -name '*_string.go'`
cleancover:
	-rm -f $(COVER_PROFILE)

.PHONY: all generate test testdeps deps cover systemctl init install clean cleanbin cleanstringers cleancover 
