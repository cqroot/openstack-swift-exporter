FROM       python:2.7.18-alpine3.11
MAINTAINER cqroot
LABEL      maintainer="cqroot <cqroot@outlook.com>"

COPY       .build/swift_exporter update_swift_info.py endpoint.sh /bin/
RUN        echo '0       *       *       *       *       python /bin/update_swift_info.py' >> /etc/crontabs/root && \
           mkdir -p /etc/swift_exporter

EXPOSE     9150
ENTRYPOINT ["sh", "/bin/endpoint.sh"]
