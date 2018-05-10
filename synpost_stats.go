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

type job struct {
	ID      bson.ObjectId `bson:"_id"`
	Created time.Time     `bson:"created"`
}

func main() {
	log.Printf("synpost_stats (v%s)", appVersion)
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

	type ReportFunc func(*mgo.Session, *statsd.Client, chan error)
	var functions = [...]ReportFunc{
		reportPendingImportJobs,
		reportQueuedImportJobs,
		reportProcessingImportJobs,
		reportBrokenScheduledAutoresponders,
		reportAverageAgeImportJobs,
		// add more functions here in future.
	}

	var ch = make(chan error)
	for {
		for _, f := range functions {
			go f(session, c, ch)
		}
		for i := 0; i < len(functions); i++ {
			err := <-ch
			if err != nil {
				panic(err)
			}
		}
		c.Flush()
		time.Sleep(time.Second * 60)
	}
}

func reportPendingImportJobs(session *mgo.Session, c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus(session, "Pending")
	if err != nil {
		log.Print("Unrecoverable error in reportPendingImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of pending import jobs: %d", n)
	c.Gauge("jobs.import.status.pending", n)
	ch <- nil
}

func reportQueuedImportJobs(session *mgo.Session, c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus(session, "Queued")

	if err != nil {
		log.Print("Unrecoverable error in reportQueuedImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of queued import jobs: %d", n)
	c.Gauge("jobs.import.status.queued", n)
	ch <- nil
}

func reportProcessingImportJobs(session *mgo.Session, c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus(session, "Processing")

	if err != nil {
		log.Print("Unrecoverable error in reportProcessingImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of processing import jobs: %d", n)
	c.Gauge("jobs.import.status.processing", n)
	ch <- nil
}

func reportBrokenScheduledAutoresponders(session *mgo.Session, c *statsd.Client, ch chan error) {
	coll := session.DB("synpost").C("Campaign")
	n, err := coll.Find(bson.M{"type": "scheduled-autoresponder", "segment": bson.M{"$exists": false}, "status": "Scheduled"}).Count()
	if err != nil {
		log.Print("Unrecoverable error in reportBrokenScheduledAutoresponders: ", err)
		ch <- err
		return
	}
	log.Printf("Number of broken scheduled autoresponders: %d", n)
	c.Gauge("campaigns.scheduled_ar.broken", n)
	ch <- nil
}

func getJobCountByStatus(session *mgo.Session, status string) (int, error) {
	coll := session.DB("synpost").C("Job")
	n, err := coll.Find(bson.M{"jobType": "Import Subscriber", "status": status}).Count()

	return n, err
}

func reportAverageAgeImportJobs(session *mgo.Session, c *statsd.Client, ch chan error) {
	var res job
	var age int
	var count int
	var average int
	coll := session.DB("synpost").C("Job")
	iter := coll.Find(bson.M{
		"jobType": "Import Subscriber",
		"status": bson.M{"$in": [3]string{
			"Pending",
			"Queued",
			"Processing",
		}},
	}).Iter()
	for iter.Next(&res) {
		age += int(time.Since(res.Created).Seconds())
		count++
	}
	if count != 0 {
		average = age / count
	}
	log.Printf("Average age of import jobs: %d", average)
	c.Gauge("jobs.import.age.average", average)
	ch <- nil
}
