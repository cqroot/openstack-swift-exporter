.PHONY: build run clean pack dbuild drun dexec dclean

build:
	CGO_ENABLED=0 go build -o bin/swift_exporter main.go

run: pack
	CGO_ENABLED=0 go build -o bin/swift_exporter main.go
	bin/swift_exporter --log.debug

clean:
	go clean
	rm -rf ./bin/swift_exporter ./swift_exporter

pack: build
	mkdir -p swift_exporter
	mkdir -p swift_exporter/bin
	mkdir -p swift_exporter/conf
	cp -r bin swift_exporter/
	cp bin/update_swift_info.py swift_exporter/bin/
	cp -r conf/ swift_exporter/

# docker
dbuild: pack
	docker build --force-rm -t swift_exporter .

drun:
	docker run \
		-itd -p 9150:9150 \
		-v /etc/swift:/etc/swift \
		--hostname swift_exporter \
		--name swift_exporter \
		swift_exporter --log.debug

dexec:
	docker exec -it swift_exporter /bin/sh

dclean: clean
	docker rm -f swift_exporter; docker rmi swift_exporter
