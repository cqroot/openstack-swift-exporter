FROM       python:2.7.18-alpine3.11
MAINTAINER cqroot
LABEL      maintainer="cqroot <cqroot@outlook.com>"

COPY       .build/swift_exporter update_swift_info.py /bin/
RUN        echo '0       *       *       *       *       python /bin/update_swift_info.py' >> /etc/crontabs/root

EXPOSE     9150
ENTRYPOINT ["/bin/swift_exporter"]
