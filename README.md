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

## Prerequisites

Local development of this project is done on Go lang and Docker.

* Install Go lang 1.9.x
* Install Docker CE (Community Edition) version 17.09 or above.
* Atom is the recommended IDE. Be sure to install `go-plus` package.

## Development

First, bring up a mongo cluster (3 mongo servers in a replica set):

```
cd <appropriate_folder>
git clone https://github.com/msound/localmongo.git
cd localmongo
docker-compose  up
```

Next, bring up the application container:

```
cd <this_project_folder>
docker-compose up
```

For local development, we are using [fresh](https://github.com/pilu/fresh). So, all you have to do is write code in your IDE and save the file. `fresh` will be watching the source folder for changes. If any file is modified, it will automatically trigger a rebuild.

You should be able to open your browser and visit `http://localhost:8000` now. This should open the graphite web interface.

Please note, _in rare cases_, if there are changes to Dockerfile.dev or docker-compose.yml, you will be required to rebuild the application container. In that case, do:

```
docker rm synpost_stats_main
docker-compose build
docker-compose up
```

## Debugging

To jump into a mongo CLI for debugging purposes, do:

```
docker ps
```

Note down the container name in the last column (lets say: `localmongo1`). Now run:

```
docker exec -it localmongo1 mongo
```

If `localmongo1` is not the primary, you will still be able to run `find` queries, but not `insert` or `update` queries. For simple queries, you can run `rs.slaveOk()` and continue on the same mongo shell. Alternatively, run `rs.status()` to find out which mongo server is the primary, then ssh into that container.

## Testing

Connect to the MongoDB cli as described above.

Now, insert or update documents. Example:

```
use synpost

db.Job.insert({
  "jobType" : "Import Subscriber",
  "status" : "Pending",
  "created" : ISODate("2018-05-03T21:56:45Z"),
  "started" : ISODate("2018-05-03T21:56:52Z"),
  "updated" : ISODate("2018-05-04T16:48:27Z")
})
```

Now watch the logs of your application container and you should see the number of pending jobs has increased by one.

You can also confirm this from graphite's web interface (http://localhost:8000)

You can also confirm this from statsd's admin port. use `telnet` to connect to localhost:8126 and type in the command `gauges`.
