[![Build Status](https://travis-ci.org/b1101/systemgo.svg?branch=master&bust=1)](https://travis-ci.org/b1101/systemgo)
[![Coverage Status](https://coveralls.io/repos/github/b1101/systemgo/badge.svg?branch=master&bust=1)](https://coveralls.io/github/b1101/systemgo?branch=master)

# Description
An init system in Go, intended to run on [Browsix](https://github.com/plasma-umass/browsix) and other Unix-like OS(GNU/Linux incl.)
# Features
* Fast and concurrent
* Handles dependencies well
* [Systemd](https://github.com/Systemd/Systemd)-compatible

_Part of [GSoC project](https://summerofcode.withgoogle.com/projects/#6227933760847872)_

# Milestones
- [ ] `init` can boot an OS (_approx 05.06.2016_)
- [ ] `systemctl` can be used to `start` or `stop` units(_approx 10.06.2016_)
- [ ] _Systemgo_ functionality fully supported by [Browsix](https://github.com/plasma-umass/browsix)(_approx 15.06.2016_)
- [ ] A demo of a web service using _systemgo_ in the context of [Browsix](https://github.com/plasma-umass/browsix) is ready(_approx 20.06.2016_)
- [ ] `init` does not depend on HTTP for communication with `systemctl`

# Progress
- [x] Logging
- [x] Dependency resolution
    - [x] Wants
    - [x] Requires
    - [x] After
    - [x] Before
- [ ] Systemctl

# Supported Systemd functionality
## Commands
- [x] start
- [x] stop
- [ ] reload
- [x] restart
- [x] status
- [ ] isolate
- [ ] list-units
- [ ] enable
- [ ] disable

## Unit types
- [ ] Service
  - [x] Simple
  - [ ] Forking
  - [x] Oneshot
- [ ] Mount
- [x] Target
- [ ] Socket
