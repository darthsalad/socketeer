// Internal package for handling database methods by 
// listening for changes and dispatching updates to clients
// with the internal websocket package.
//
// This package is used in the following way:
//
// 	1. Create a new DB type with Connect().
// 	2. Listen for changes with Listen().
// 	3. Disconnect from the database with Disconnect().
//
// No need to call these methods exclusively, they are
// automatically called and are executed synchronously
// in the socketeer.go file.
package db

import (
	"encoding/json"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/darthsalad/socketeer/internal/ws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB is an interface for handling database methods.
//
// 	- Client is a mongo client.
// 	- DB is a mongo database.
// 	- Coll is a mongo collection.
type DB struct {
	Client *mongo.Client
	DB     *mongo.Database
	Coll   *mongo.Collection
}

// UpdateEvent is a struct for handling 
// mongo update events from the database.
//
// 	- OperationType is the type of operation,
// 		which is always "update".
// 	- UpdateDescription is a struct for handling
// 		the updated fields.
type UpdateEvent struct {
	OperationType     string `bson:"operationType"`
	UpdateDescription struct {
		UpdatedFields bson.M `bson:"updatedFields"`
	} `bson:"updateDescription"`
}

// CreateEvent is a struct for handling
// mongo create events from the database.
//
// 	- OperationType is the type of operation,
// 		which is always "insert".
// 	- FullDocument is a struct for handling
// 		the full document.
type CreateEvent struct {
	OperationType string `bson:"operationType"`
	FullDocument  bson.M `bson:"fullDocument"`
}

// Connect returns a new DB type by
// connecting to the database with the uri,
// database name, and collection name provided.
//
// This method is utilized to create a new DB type
// and is called internally when the socketeer is started.
//
// # Parameters:
//
// 	- uriString (string): the uri string to connect to the database, example: mongodb://localhost:27017
// 	- dbName (string): the name of the database to connect to, example: mydb
// 	- collName (string): the name of the collection to connect to, example: mycollection
//
// # Example:
//
// 	db.Connect("mongodb://localhost:27017", "mydb", "mycollection")
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

// Listen listens for changes in the database
// by the mongo watch & changeStream methods and dispatches updates
// to clients with the internal websocket package.
//
// This method is called internally when the socketeer is started.
//
// # Parameters:
//
// 	- ws (WebSocket): the WebSocket type to dispatch updates to.
// 	- keys ([]string): the keys in the documents of the collection 
// 		to listen for changes on.
//
// # Example:
//
// 	db.Listen(ws, []string{"displayName", "email"})
func (d *DB) Listen(ws *ws.WebSocket, keys []string) error {
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
			var responseMap = make(map[string]string)
			fmt.Println("Update event")
			for key, value := range updateResult.UpdateDescription.UpdatedFields {
				for _, k := range keys {
					if key == k {
						responseMap[key] = fmt.Sprintf("%v", value)
					}
				}
			}
			data, err := json.Marshal(responseMap)
			if err != nil {
				log.Fatal(err)
				return err
			}
			ws.DispatchUpdate(data)
		} else if createResult.OperationType == "insert" {
			fmt.Println("Create event")
			var responseMap = make(map[string]string)
			for key, value := range createResult.FullDocument {
				for _, k := range keys {
					if key == k {
						responseMap[key] = fmt.Sprintf("%v", value)
					}
				}
			}
			data, err := json.Marshal(responseMap)
			if err != nil {
				log.Fatal(err)
				return err
			}
			ws.DispatchUpdate(data)
		}
	}

	return nil
}

// Disconnect ends the connection to the database.
//
// This method is called internally when the socketeer is stopped.
//
// # Example:
//
// 	db.Disconnect()
func (d *DB) Disconnect() error {
	err := d.Client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
