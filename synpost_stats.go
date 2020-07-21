package main

import (
	"context"
	"log"
	"time"

	"github.com/synduit/synpost_stats/synmongo"
	"github.com/synduit/synpost_stats/synstatsd"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"gopkg.in/alexcesaro/statsd.v2"
)

const (
	jobCollection           = "Job"
	campaignCollection      = "Campaign"
	jobTypeImportSubscriber = "Import Subscriber"
)

type job struct {
	ID      primitive.ObjectID `bson:"_id"`
	Created time.Time          `bson:"created"`
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

	log.Print("Creating statsd client")
	c := synstatsd.GetStatsd()
	defer c.Close()

	type ReportFunc func(*statsd.Client, chan error)
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
			go f(c, ch)
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

func reportPendingImportJobs(c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus("Pending")
	if err != nil {
		log.Print("Unrecoverable error in reportPendingImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of pending import jobs: %d", n)
	c.Gauge("jobs.import.status.pending", n)
	ch <- nil
}

func reportQueuedImportJobs(c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus("Queued")

	if err != nil {
		log.Print("Unrecoverable error in reportQueuedImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of queued import jobs: %d", n)
	c.Gauge("jobs.import.status.queued", n)
	ch <- nil
}

func reportProcessingImportJobs(c *statsd.Client, ch chan error) {
	n, err := getJobCountByStatus("Processing")

	if err != nil {
		log.Print("Unrecoverable error in reportProcessingImportJobs: ", err)
		ch <- err
		return
	}

	log.Printf("Number of processing import jobs: %d", n)
	c.Gauge("jobs.import.status.processing", n)
	ch <- nil
}

func reportBrokenScheduledAutoresponders(c *statsd.Client, ch chan error) {
	mdb, con := synmongo.GetMongoConnection()
	defer con.Disconnect(context.Background())

	col := mdb.Collection(campaignCollection)

	filter := bson.M{}
	filter["type"] = "scheduled-autoresponder"
	filter["segment"] = bson.M{"$exists": false}
	filter["status"] = "Scheduled"

	n, err := col.CountDocuments(context.TODO(), filter)

	if err != nil {
		log.Print("Unrecoverable error in reportBrokenScheduledAutoresponders: ", err)
		ch <- err
		return
	}
	log.Printf("Number of broken scheduled autoresponders: %d", n)
	c.Gauge("campaigns.scheduled_ar.broken", n)
	ch <- nil
}

func getJobCountByStatus(status string) (int, error) {
	mdb, con := synmongo.GetMongoConnection()
	defer con.Disconnect(context.Background())

	c := mdb.Collection(jobCollection)

	filter := bson.M{}
	filter["jobType"] = jobTypeImportSubscriber
	filter["status"] = status

	n, err := c.CountDocuments(context.TODO(), filter)

	return int(n), err
}

func reportAverageAgeImportJobs(c *statsd.Client, ch chan error) {
	var age int
	var count int
	var average int
	var cur *mongo.Cursor

	mdb, con := synmongo.GetMongoConnection()
	defer con.Disconnect(context.Background())

	col := mdb.Collection(jobCollection)

	filter := bson.M{}
	filter["jobType"] = jobTypeImportSubscriber
	filter["status"] = bson.M{"$in": [3]string{
		"Pending",
		"Queued",
		"Processing",
	}}

	cur, _ = col.Find(context.TODO(), filter)

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		job := &job{}
		err := cur.Decode(job)

		if err == nil {
			age += int(time.Since(job.Created).Seconds())
			count++
		}
	}

	if count != 0 {
		average = age / count
	}
	log.Printf("Average age of import jobs: %d", average)
	c.Gauge("jobs.import.age.average", average)
	ch <- nil
}
