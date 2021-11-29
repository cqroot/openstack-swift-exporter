#!/bin/bash

/usr/sbin/crond
python /bin/update_swift_info.py
"/bin/swift_exporter" $@
