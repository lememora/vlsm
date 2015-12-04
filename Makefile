default: test
build:
	go install github.com/lememora/vlsm
test: build
	vlsm