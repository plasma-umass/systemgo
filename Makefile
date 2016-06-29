COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html
TEST_LIB=lib/test

all: build test

deps:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/coreos/go-systemd/unit
	go get -d ./...
testdeps: mockdeps
mockdeps:
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
coverdeps:
	go get golang.org/x/tools/cmd/cover
	go get github.com/wadey/gocovmerge
coveralldeps: coverdeps
	go get github.com/mattn/goveralls

build: generate init systemctl

generate: deps
	go generate ./...
init: *.go system/*.go unit/*.go lib/test/*.go lib/errors/*.go generate
	go build -o bin/init
systemctl: systemctl/cmd/*.go systemctl/main.go lib/systemctl/*.go generate
	go build -o bin/systemctl ./systemctl
install: build
	go get -v ./...

testall: cover

mock: mockdeps
	mkdir -p $(TEST_LIB)/mock_system
	mkdir -p $(TEST_LIB)/mock_unit
	mockgen -destination=$(TEST_LIB)/mock_system/mock_system.go -package=mock_system github.com/b1101/systemgo/system Supervisable,Daemon,Manager
	mockgen -destination=$(TEST_LIB)/mock_unit/mock_unit.go -package=mock_unit github.com/b1101/systemgo/unit Starter,Stopper,StartStopper,Reloader,Subber
	go get ./$(TEST_LIB)/...
test: testdeps generate mock 
	go test ./...
cover: coverdeps generate
	./coverage.sh --html
coveralls: coveralldeps generate 
	./coverage.sh --coveralls
travis: coveralldeps build
	./coverage.sh --coveralls travis-ci

clean: cleancover cleanbin cleanstringers cleanmock

cleanbin:
	-rm -rf bin
cleanstringers:
	-rm `find -name '*_string.go'`
cleancover:
	-rm -f $(COVER_PROFILE)
cleanmock:
	-rm -rf `find -name 'mock_*'`

.PHONY: all generate test testdeps deps cover coverdeps systemctl init install clean cleanbin cleanstringers cleancover cleanmock build travis
