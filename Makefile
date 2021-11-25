BINARY_NAME=swift_exporter

build:
	CGO_ENABLED=0 go build -o .build/${BINARY_NAME} main.go

run:
	CGO_ENABLED=0 go build -o .build/${BINARY_NAME} main.go
	./.build/${BINARY_NAME} -debug

clean:
	go clean
	rm -r ./.build/

dbuild: build
	docker build --force-rm -t swift_exporter .

drun:
	docker run \
		-itd -p 9150:9150 \
		-v /etc/swift:/etc/swift \
		--hostname swift_exporter \
		--name swift_exporter \
		swift_exporter -debug

dexec:
	docker exec -it swift_exporter /bin/sh

dclean:
	docker rm -f swift_exporter; docker rmi swift_exporter
