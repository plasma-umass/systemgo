REPO=github.com/rvolosatovs/systemgo

TEST=test
UNIT=unit
SYSTEM=system
SYSTEMCTL=cmd/systemctl
INIT=cmd/init

BINDIR=bin

COVER=cover.sh
COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html
TRAVIS_MODE=travis-ci

PKGS=`go list $(REPO)/... | grep -v "vendor"`

PKG_SYSTEMGO=$(REPO)
PKG_TEST=$(REPO)/$(TEST)
PKG_UNIT=$(REPO)/$(UNIT)
PKG_INIT=$(REPO)/$(INIT)
PKG_SYSTEM=$(REPO)/$(SYSTEM)
PKG_SYSTEMCTL=$(REPO)/$(SYSTEMCTL)


ABS_REPO=$(GOPATH)/src/$(REPO)

ABS_BINDIR=$(ABS_REPO)/bin

ABS_COVER=$(ABS_REPO)/cover.sh
ABS_COVER_PROFILE=$(ABS_REPO)/cover.out

ABS_SYSTEMGO=$(ABS_REPO)
ABS_TEST=$(ABS_REPO)/$(TEST)
ABS_UNIT=$(ABS_REPO)/$(UNIT)
ABS_INIT=$(ABS_REPO)/$(INIT)
ABS_SYSTEM=$(ABS_REPO)/$(SYSTEM)
ABS_SYSTEMCTL=$(ABS_REPO)/$(SYSTEMCTL)

MOCK_PKGS=mock_unit mock_systemctl
#system_interfaces=Supervisable,Dependency,Reloader
unit_interfaces=Interface,Reloader,Starter,Stopper
systemctl_interfaces=Daemon

all: build test

build: generate vet init systemctl

depend:
	@echo "Checking build dependencies..."
	@go get -v golang.org/x/tools/cmd/stringer
	@go get -v github.com/coreos/go-systemd/unit
	@go get -v -d $(REPO)/...

dependtest: dependmock
	@go get -v github.com/stretchr/testify

dependmock:
	@echo "Checking mock testing dependencies..."
	@go get -v github.com/golang/mock/gomock
	@go get -v github.com/golang/mock/mockgen

dependcover: dependtest
	@echo "Checking coverage testing dependencies..."
	@go get -v golang.org/x/tools/cmd/cover
	@go get -v github.com/wadey/gocovmerge

dependcoverall: dependcover
	@echo "Checking coveralls.io testing dependencies..."
	@go get -v github.com/mattn/goveralls

vet: generate
	@echo "Running 'go vet'..."
	@go vet $(PKGS)

generate: depend
	@echo "Running 'go generate'..."
	@go generate -x $(REPO)/...

$(MOCK_PKGS) test cover build: generate

init systemctl: % : $(wildcard cmd/%/*.go)
	@echo "Building $@..."
	@go build -o $(ABS_BINDIR)/$@ $(REPO)/cmd/$@
	@echo "$@ built and saved to $(ABS_BINDIR)/$@"

install: build
	@echo "Installing..."
	@go get -v $(REPO)/...


mock: dependmock $(MOCK_PKGS)

$(MOCK_PKGS): mock_%: $(wildcard %/interfaces.go)
	@echo "Mocking $* interfaces..."
	@mkdir -p $(ABS_TEST)/$@
	@mockgen -destination=$(ABS_TEST)/$@/$@.go -package=$@ $(REPO)/$* $($*_interfaces)
	@echo "$@ package built and saved to $(ABS_TEST)/$@"
	@go get $(PKG_TEST)/$@

test: dependtest mock 
	@echo "Starting tests..."
	@go test -v $(PKGS)

cover: dependcover
	@echo "Creating html coverage report..."
	@$(ABS_COVER) --html

coveralls: dependcoverall
	@echo "Pushing coverage statistics to coveralls.io..."
	@$(ABS_COVER) --coveralls


travis: dependcoverall build mock
	@echo "Starting travis build..."
	@$(ABS_COVER) --coveralls $(TRAVIS_MODE)


clean: cleancover cleanbin cleanstringers cleanmock

cleanbin:
	@echo "Removing compiled binaries..."
	@-rm -rf $(ABS_BINDIR)
cleanstringers:
	@echo "Removing generated stringers..."
	@-rm `find $(ABS_REPO) -name '*_string.go'`
cleancover:
	@echo "Removing coverage profile..."
	@-rm -f $(ABS_COVER_PROFILE)
cleanmock:
	@echo "Removing mock units..."
	@-rm -rf `find $(ABS_REPO) -name 'mock_*'`

fix:
	@echo "rvolosatovs -> rvolosatovs"
	@./fix-import.sh

.PHONY: all generate test dependtest depend cover dependcover systemctl init install clean cleanbin cleanstringers cleancover cleanmock build travis mock_system mock_unit vet init cmd/init cmd/systemctl $(SYSTEMCTL)
