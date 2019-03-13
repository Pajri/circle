package utils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConfigDB(ctx context.Context) (*mongo.Database, error) {
	username := ""
	password := ""
	host := ""
	database := ""

	uri := fmt.Sprintf("mongodb:%s://%s@%s/%s", username, password, host, database)

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("db utils: couldn't connect to mongo: %v", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("db utils: mongo client couldn't connect with background context: %v", err)
	}

	return client.Database(database), nil
}
