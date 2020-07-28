package synmongo

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoConnectTimeout     = 20 // in seconds
	mongoIdleConnectTimeout = 20 // in milliseconds
)

// Mongo represents db for connecting
type Mongo struct{}

// Connect connects to mongodb.
func (Mongo) Connect() (interface{}, error) {
	connectionURI, present := os.LookupEnv("MONGODB_URL")
	if !present {
		log.Fatal("Need to set MONGODB_URL environment variable")
	}

	poolSize, present := os.LookupEnv("DATABASE_POOL_SIZE")
	if !present {
		log.Fatal("Need to set DATABASE_POOL_SIZE environment variable")
	}
	poolSizeInt, _ := strconv.Atoi(poolSize)

	ctx, cancel := context.WithTimeout(context.Background(), mongoConnectTimeout*time.Second)
	defer cancel()
	clientOptions := options.Client().SetMaxConnIdleTime(mongoIdleConnectTimeout * time.Millisecond).SetMaxPoolSize(uint64(poolSizeInt)).ApplyURI(connectionURI)
	client, err := mongo.Connect(ctx, clientOptions)

	return client, err
}

// GetMongoConnection helper function to get the db connection.
func GetMongoConnection() (*mongo.Database, *mongo.Client) {
	var m Mongo
	s, err := m.Connect()
	if err != nil {
		log.Panic("Unrecoverable error in connecting mongo: ", err)
	}

	client := s.(*mongo.Client)
	dbName, present := os.LookupEnv("DATABASE_NAME")
	if !present {
		log.Fatal("Need to set DATABASE_NAME environment variable")
	}
	mdb := client.Database(dbName)

	return mdb, client
}
