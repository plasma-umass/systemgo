REPO=github.com/b1101/systemgo

TEST=lib/test
UNIT=unit
SYSTEM=system
SYSTEMCTL=systemctl

BINDIR=bin

COVER=cover.sh
COVER_PROFILE=cover.out
COVER_METHOD=atomic
COVER_DEFAULT=html
TRAVIS_MODE=travis-ci


PKG_SYSTEMGO=$(REPO)
PKG_TEST=$(REPO)/$(TEST)
PKG_UNIT=$(REPO)/$(UNIT)
PKG_SYSTEM=$(REPO)/$(SYSTEM)
PKG_SYSTEMCTL=$(REPO)/$(SYSTEMCTL)


ABS_REPO=$(GOPATH)/src/$(REPO)

ABS_BINDIR=$(ABS_REPO)/bin

ABS_COVER=$(ABS_REPO)/cover.sh
ABS_COVER_PROFILE=$(ABS_REPO)/cover.out

ABS_SYSTEMGO=$(ABS_REPO)
ABS_TEST=$(ABS_REPO)/$(TEST)
ABS_UNIT=$(ABS_REPO)/$(UNIT)
ABS_SYSTEM=$(ABS_REPO)/$(SYSTEM)
ABS_SYSTEMCTL=$(ABS_REPO)/$(SYSTEMCTL)

mock_units=mock_system mock_unit

all: build test

depend:
	@echo "Checking build dependencies..."
	@go get -v golang.org/x/tools/cmd/stringer
	@go get -v github.com/coreos/go-systemd/unit
	@go get -v -d ./...

dependtest: dependmock

dependmock:
	@echo "Checking mock testing dependencies..."
	@go get -v github.com/golang/mock/gomock
	@go get -v github.com/golang/mock/mockgen

dependcover:
	@echo "Checking coverage testing dependencies..."
	@go get -v golang.org/x/tools/cmd/cover
	@go get -v github.com/wadey/gocovmerge

dependcoverall: dependcover
	@echo "Checking coveralls.io testing dependencies..."
	@go get -v github.com/mattn/goveralls


build: generate init systemctl

generate: depend
	@echo "Running 'go generate'..."
	@go generate -x $(REPO)/...

$(mock_units) test mock cover build: generate

init: *.go system/*.go unit/*.go lib/systemctl/*.go
	@echo "Building $@..."
	@go build -o $(ABS_BINDIR)/$@ $(REPO)
	@echo "$@ built and saved to $(ABS_BINDIR)/$@"

systemctl: systemctl/cmd/*.go systemctl/main.go lib/systemctl/*.go 
	@echo "Building $@..."
	@go build -o $(ABS_BINDIR)/$@ $(REPO)/$@
	@echo "$@ built and saved to $(ABS_BINDIR)/$@"

install: build
	@echo "Installing..."
	@go get -v $(REPO)/...

mock: dependmock mock_system mock_unit
	@go get $(PKG_TEST)/...

mock_system: system/interfaces.go
	@echo "Building $@..."
	@mkdir -p $(ABS_TEST)/$@
	@mockgen -destination=$(ABS_TEST)/$@/$@.go -package=$@ $(PKG_SYSTEM) Supervisable,Daemon,Manager
	@echo "$@ package built and saved to $(ABS_PATH)/$(ABS_TEST)/$@"

mock_unit: unit/interfaces.go
	@echo "Building $@..."
	@mkdir -p $(ABS_TEST)/$@
	@mockgen -destination=$(ABS_TEST)/$@/$@.go -package=$@ $(PKG_UNIT) Starter,Stopper,StartStopper,Reloader,Subber
	@echo "$@ package built and saved to $(ABS_TEST)/$@"

test: dependtest mock 
	@echo "Starting tests..."
	@go test -v $(REPO)/...

cover: dependcover
	@echo "Creating html coverage report..."
	@$(ABS_COVER) --html

coveralls: dependcoverall
	@echo "Pushing coverage statistics to coveralls.io..."
	@$(ABS_COVER) --coveralls


travis: dependcoverall build
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

.PHONY: all generate test dependtest depend cover dependcover systemctl init install clean cleanbin cleanstringers cleancover cleanmock build travis mock_system mock_unit
