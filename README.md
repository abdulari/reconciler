# reconciler
A CLI reconciliation loop engine written in Go.


## example usage 
``` bash
reconciler -config config.yaml -output logfmt -every 60 -log-dir /var/log/reconciler

```


## example usage with systemd

``` bash
# /etc/systemd/system/reconciler.service
[Unit]
Description=Reconciler

[Service]
ExecStart=/usr/bin/reconciler -config /etc/reconciler/config.yaml -output logfmt -every 60 -log-dir /var/log/reconciler
Restart=always
RestartSec=5

```