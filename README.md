# OpenStack Swift Exporter

![](https://img.shields.io/github/go-mod/go-version/cqroot/openstack_swift_exporter)
![](https://img.shields.io/github/v/release/cqroot/openstack_swift_exporter)

OpenStack Swift Exporter for Prometheus.

## Installation and Usage

### Systemd

```bash
make pack
cp systemd/swift-exporter.service /usr/lib/systemd/system/
mv swift_exporter /opt/

systemctl start swift-exporter.service
systemctl enable swift-exporter.service
```

### Docker

```bash
make docker-build
make docker-run
```

## Collectors

You can use http parameters to filter the collector used:

```
http://127.0.0.1:9150/metrics?collect=disk&collect=server
```

Available collectors:

| collector | scrape speed |
|-----------|--------------|
| server    | fast         |
| proxy     | fast         |
| disk      | slow         |

Default is ["server"].

## Metrics

### Server

| Name                          | Description                          |
| :---------------------------- | :----------------------------------- |
| swift_server_account_status   | Swift account-server reachability.   |
| swift_server_container_status | Swift container-server reachability. |
| swift_server_object_status    | Swift object-server reachability.    |

### Proxy

| Name                      | Description                                    |
|:--------------------------|:-----------------------------------------------|
| swift_proxy_put_status    | Swift proxy-server put request test status.    |
| swift_proxy_delete_status | Swift proxy-server delete request test status. |

### Disk

| Name                         | Description |
|:-----------------------------|:------------|
| swift_disk_avail_bytes       |             |
| swift_disk_used_bytes        |             |
| swift_disk_size_bytes        |             |
| swift_disk_usage_bytes       |             |
| swift_disk_total_avail_bytes |             |
| swift_disk_total_used_bytes  |             |
| swift_disk_total_size_bytes  |             |
