.PHONY: build
build:
	@BuildVersion=$$(git describe --tags --abbrev=0); \
	  echo "Build Version: $${BuildVersion}"; \
	  sed -i "s/BuildVersion string = \"[^\"]*\"/BuildVersion string = \"$${BuildVersion}\"/" internal/version.go
	@CGO_ENABLED=0 go build -o bin/swift_exporter main.go

.PHONY: run
run: pack
	@echo ''
	@swift_exporter/bin/swift_exporter --config swift_exporter/conf/swift_exporter.yml

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/swift_exporter ./swift_exporter

.PHONY: pack
pack: build
	mkdir -p swift_exporter
	mkdir -p swift_exporter/bin
	mkdir -p swift_exporter/conf
	cp -r bin systemd swift_exporter/
	cp bin/update_swift_info.py swift_exporter/bin/
	cp -r conf/ swift_exporter/

.PHONY: docker-build
docker-build: pack
	docker build --force-rm -t swift_exporter .

.PHONY: docker-run
docker-run:
	docker run \
		-itd -p 9150:9150 \
		-v /etc/swift:/etc/swift \
		--hostname swift_exporter \
		--name swift_exporter \
		swift_exporter --log.debug

.PHONY: docker-exec
docker-exec:
	docker exec -it swift_exporter /bin/sh

.PHONY: docker-clean
docker-clean: clean
	docker rm -f swift_exporter; docker rmi swift_exporter
