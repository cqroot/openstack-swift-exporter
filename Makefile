BINARY_NAME=swift_exporter
OS=linux
ARCH=amd64

build:
	CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -o .build/${OS}-${ARCH}/${BINARY_NAME} main.go

run:
	CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -o .build/${OS}-${ARCH}/${BINARY_NAME} main.go
	./.build/${OS}-${ARCH}/${BINARY_NAME}

clean:
	go clean
	rm ./.build/
