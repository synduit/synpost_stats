# synpost_stats
A StatsD monitoring tool for Synpost application metrics

Expected ENV variables
* SYNPOST_MONGO_SERVER (ex: 192.168.33.95)
* STATSD_HOST (ex: 192.168.99.101)
* STATSD_PORT (ex: 8125)

If you want to run the binary in Aplinx linux, when compiling, please be sure to do:
```
export CGO_ENABLED=0
```
