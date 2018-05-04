package synstatsd

import (
	"log"
	"os"

	"gopkg.in/alexcesaro/statsd.v2"
)

// GetStatsd returns as statsd client.
func GetStatsd() *statsd.Client {
	statsdServer := os.Getenv("STATSD_HOST")
	statsdPort := os.Getenv("STATSD_PORT")
	if statsdPort == "" {
		statsdPort = "8125"
	}
	c, err := statsd.New(
		statsd.Address(statsdServer+":"+statsdPort),
		statsd.Prefix("synpost"),
	)
	if err != nil {
		log.Panic("Unrecoverable error in getStatsd: ", err)
	}

	return c
}
