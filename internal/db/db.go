package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Client *mongo.Client
	DB     *mongo.Database
	Coll   *mongo.Collection
}

type UpdateEvent struct {
	OperationType     string `bson:"operationType"`
	UpdateDescription struct {
		UpdatedFields bson.M `bson:"updatedFields"`
	} `bson:"updateDescription"`
}

type CreateEvent struct {
	OperationType string `bson:"operationType"`
	FullDocument  bson.M `bson:"fullDocument"`
}

func Connect(uriString string, dbName string, collName string) (*DB, error) {
	clientOptions := options.Client().ApplyURI(uriString).SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags: true,
	})

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &DB{
		Client: client,
		DB:     client.Database(dbName),
		Coll:   client.Database(dbName).Collection(os.Getenv(collName)),
	}, nil
}

func (d *DB) Listen() error {
	coll := d.Coll
	changeStream, err := coll.Watch(context.Background(), mongo.Pipeline{}, options.ChangeStream())
	if err != nil {
		log.Fatal(err)
		return err
	}

	for changeStream.Next(context.Background()) {
		var updateResult UpdateEvent
		var createResult CreateEvent
		var temp bson.D
		err := changeStream.Decode(&temp)
		if err != nil {
			log.Fatal(err)
			return err
		}

		for _, item := range temp {
			if item.Key == "operationType" {
				if item.Value == "update" {
					updateResult = UpdateEvent{}
					bsonBytes, err := bson.Marshal(temp)
					if err != nil {
						log.Fatal(err)
						return err
					}
					bson.Unmarshal(bsonBytes, &updateResult)
				} else if item.Value == "insert" {
					createResult = CreateEvent{}
					bsonBytes, err := bson.Marshal(temp)
					if err != nil {
						log.Fatal(err)
						return err
					}
					bson.Unmarshal(bsonBytes, &createResult)
				}
			}
		}

		if updateResult.OperationType == "update" {
			fmt.Println("Update event")
			for key, value := range updateResult.UpdateDescription.UpdatedFields {
				fmt.Println("Key: ", key, "\t Value: ", value)
			}
		} else if createResult.OperationType == "insert" {
			fmt.Println("Create event")
			for key, value := range createResult.FullDocument {
				fmt.Println("Key: ", key, "\t Value: ", value)
			}
		}
	}

	return nil
}

func (d *DB) Disconnect() error {
	err := d.Client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
