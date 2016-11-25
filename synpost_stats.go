package main

import (
	"log"
	"time"

	"github.com/synduit/synpost_stats/synmongo"
	"github.com/synduit/synpost_stats/synstatsd"

	"gopkg.in/alexcesaro/statsd.v2"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	log.Print("Sleeping for few seconds until DNS entries are stabilized")
	time.Sleep(time.Second * 2)

	for {
		reportStats()
		log.Print("Sleeping for 30 seconds before retrying")
		time.Sleep(time.Second * 30)
	}
}

func reportStats() {
	defer func() {
		if err := recover(); err != nil {
			log.Print("Recovering from panic: ", err)
		}
	}()

	log.Print("Connecting to mongo")
	session := synmongo.GetMongo()
	defer session.Close()

	log.Print("Creating statsd client")
	c := synstatsd.GetStatsd()
	defer c.Close()

	var ch = make(chan error)
	var err error
	for {
		go reportPendingImportJobs(session, c, ch)
		err = <-ch
		c.Flush()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 10)
	}
}

func reportPendingImportJobs(session *mgo.Session, c *statsd.Client, ch chan error) {
	coll := session.DB("synpost").C("Job")
	n, err := coll.Find(bson.M{"jobType": "Import Subscriber", "status": "Pending"}).Count()
	if err != nil {
		log.Print("Unrecoverable error in reportPendingImportJobs: ", err)
		ch <- err
	} else {
		log.Printf("Number of import jobs: %d", n)
		c.Gauge("jobs.import_pending", n)
		ch <- nil
	}
}
