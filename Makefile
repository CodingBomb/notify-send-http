.PHONY: gox

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OS = linux
endif
ifeq ($(UNAME_S),Darwin)
	OS = darwin
endif

default: xcompile

gox:
	go get -v github.com/mitchellh/gox
	gox -verbose -build-toolchain -osarch="darwin/amd64 linux/amd64"

xcompile: client/main.go server/main.go
	gox -verbose -osarch="darwin/amd64 linux/amd64" -output="build/notify-client-{{.OS}}_{{.Arch}}" ./client
	gox -verbose -osarch="darwin/amd64 linux/amd64" -output="build/notify-server-{{.OS}}_{{.Arch}}" ./server
	chmod +x build/*

PORT = 13579

listen:
	PORT=${PORT} ./build/notify-server-${OS}_amd64 \

listen-d:
	PORT=${PORT} nohup ./build/notify-server-${OS}_amd64 > listen.log 2>&1 &

kill:
	kill $(shell ps aux | grep notify-server-${OS}_amd64 | grep -v grep | awk '{print $$2}')

test:
	NOTIFY_SEND_URL="http://127.0.0.1:${PORT}" ./build/notify-client-${OS}_amd64 "Test Notification" "Nice! This is a test notification." --icon icon.png