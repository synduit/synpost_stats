package synmongo

import (
	"log"
	"os"

	"gopkg.in/mgo.v2"
)

// GetMongo returns a MongoDB session.
func GetMongo() *mgo.Session {
	mongoServer := os.Getenv("SYNPOST_MONGO_SERVER")
	session, err := mgo.Dial(mongoServer)
	if err != nil {
		log.Panic("Unrecoverable error in getMongo: ", err)
	}

	return session
}
