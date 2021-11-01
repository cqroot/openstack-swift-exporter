BINARY_NAME=swift_exporter

build:
	CGO_ENABLED=0 go build -o .build/${BINARY_NAME} main.go

run:
	CGO_ENABLED=0 go build -o .build/${BINARY_NAME} main.go
	./.build/${BINARY_NAME} -debug

clean:
	go clean
	rm -r ./.build/
