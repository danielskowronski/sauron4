# sauron4
fourth approach to **Sauron - *a real time eye on your network*** - this time in Go Language and optionally sending data to InfluxDB

## features
* ICMP ping to listed targets
* optional sending of results to InfluxDB

## execution
```bash
go run sauron4.go example_config.yml
```
```
local_hostname_google: loss 0.00%, rtt 10ms
local_hostname_cloudflare: loss 0.00%, rtt 11ms
```

## config
* for `pinger_params.ICMP` refer to https://github.com/go-ping/ping/blob/master/ping.go
