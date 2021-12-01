.PHONY: build run clean pack dbuild drun dexec dclean

build:
	CGO_ENABLED=0 go build -o .build/swift_exporter main.go

run:
	CGO_ENABLED=0 go build -o .build/swift_exporter main.go
	./.build/swift_exporter -debug

clean:
	go clean
	rm -r ./.build/

pack: build
	mkdir -p swift_exporter
	mkdir -p swift_exporter/bin
	mkdir -p swift_exporter/conf
	cp .build/swift_exporter swift_exporter/bin/
	cp bin/update_swift_info.py swift_exporter/bin/
	cp -r conf/ swift_exporter/

# docker
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
