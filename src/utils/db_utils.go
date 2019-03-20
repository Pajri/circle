package utils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbUtil struct {
	username string
	password string
	host     string
	database string
	context  context.Context
}

func (db DbUtil) createURI() string {
	db = DbUtil{
		username: "",
		password: "",
		host:     "",
		database: "circle",
	}

	uri := fmt.Sprintf("mongodb:%s://%s@%s/%s", db.username, db.password, db.host, db.database)
	return uri
}

func (db DbUtil) Connect(ctx context.Context) (*mongo.Database, error) {
	uri := db.createURI()

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("db utils: couldn't connect to mongo: %v", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("db utils: mongo client couldn't connect with background context: %v", err)
	}

	return client.Database(db.database), nil
}
