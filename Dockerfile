FROM       python:2.7.18-alpine3.11
MAINTAINER cqroot
LABEL      maintainer="cqroot <cqroot@outlook.com>"

COPY       swift_exporter/ /opt/swift_exporter/

EXPOSE     9150
ENTRYPOINT ["/opt/swift_exporter/bin/swift_exporter"]
