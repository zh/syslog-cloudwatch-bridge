.PHONY: build osx clean

all: build

build: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o syslog-cloudwatch-bridge -a -tags netgo -ldflags '-w' .

osx:
	go build -o syslog-cloudwatch-osx

clean:
	rm -f syslog-cloudwatch-*
