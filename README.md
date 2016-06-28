# Description
An init system in Go, intended to run on [Browsix](https://github.com/plasma-umass/browsix) and other Unix-like OS(GNU/Linux incl.)
# Features
* Fast and concurrent
* Handles dependencies well
* [Systemd](https://github.com/Systemd/Systemd)-compatible

Part of [GSoC project](https://summerofcode.withgoogle.com/projects/#6227933760847872)

[![Build Status](https://travis-ci.org/b1101/systemgo.svg?branch=master)](https://travis-ci.org/b1101/systemgo)
[![Coverage Status](https://coveralls.io/repos/github/b1101/systemgo/badge.svg?branch=master)](https://coveralls.io/github/b1101/systemgo?branch=master)

# Progress
- [x] Logging
    - [ ] Log destination can be configured
- [x] Dependency resolution
    - [x] Wants
    - [x] Requires
    - [x] After
    - [ ] Before
- [ ] Init can boot an OS
- [ ] Systemctl can be used for communication with `init`
- [ ] Communication with `init` possible by other means than HTTP

## Commands
- [x] start
- [x] stop
- [ ] reload
- [x] restart
- [x] status
- [ ] isolate
- [ ] list-units
- [ ] daemon-reload

## Unit types
- [ ] Service
  - [x] Simple
  - [ ] Forking
  - [ ] Oneshot
- [ ] Mount
- [ ] Target
- [ ] Socket
