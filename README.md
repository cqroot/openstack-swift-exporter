# OpenStack Swift Exporter

OpenStack Swift Exporter for Prometheus.

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
| swift_account_server_status   | Swift account-server reachability.   |
| swift_container_server_status | Swift container-server reachability. |
| swift_object_server_status    | Swift object-server reachability.    |

### Proxy

| Name                | Description                                    |
|:--------------------|:-----------------------------------------------|
| swift_put_status    | Swift proxy-server put request test status.    |
| swift_delete_status | Swift proxy-server delete request test status. |

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
