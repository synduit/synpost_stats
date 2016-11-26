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
	var err1 error
	var err2 error
	for {
		go reportPendingImportJobs(session, c, ch)
		go reportBrokenScheduledAutoresponders(session, c, ch)
		err1 = <-ch
		err2 = <-ch
		c.Flush()
		if err1 != nil {
			panic(err1)
		}
		if err2 != nil {
			panic(err2)
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
		log.Printf("Number of pending import jobs: %d", n)
		c.Gauge("jobs.import_pending", n)
		ch <- nil
	}
}

func reportBrokenScheduledAutoresponders(session *mgo.Session, c *statsd.Client, ch chan error) {
	coll := session.DB("synpost").C("Campaign")
	n, err := coll.Find(bson.M{"type": "scheduled-autoresponder", "segment": bson.M{"$exists": false}}).Count()
	if err != nil {
		log.Print("Unrecoverable error in reportBrokenScheduledAutoresponders: ", err)
		ch <- err
	} else {
		log.Printf("Number of broken scheduled autoresponders: %d", n)
		c.Gauge("campaigns.scheduled_ar.broken", n)
		ch <- nil
	}
}
