// Socketeer is a package that provides a simple way to
// listen to changes in a MongoDB collection and broadcast
// them to a WebSocket server.
//
// This package is used in the following way:
//
// 	1. Create a new Socketeer type with NewSocketeer().
// 	2. Start the Socketeer with Start().
// 	3. Stop the Socketeer with Stop().
//
// These methods are to be called exclusively as per 
// the requirements of the implementation and needs.
//
// # Usage:
//
// Import the package in your project:
//
// 	import "github.com/darthsalad/socketeer"
//
// For a complete implementation, see the main.go file in the example directory.
package socketeer

import (
	"fmt"
	"log"

	"github.com/darthsalad/socketeer/internal/db"
	"github.com/darthsalad/socketeer/internal/ws"
)

// Socketeer is the main type of the package.
// It contains a pointer to a DB(internal/db.go) type and a pointer
// to a WebSocket(internal/ws.go) type.
type Socketeer struct {
	DB *db.DB
	WS *ws.WebSocket
}

// Version and Build are the version and build of the package.
var (
	Version = "1.0.1"
)

// NewSocketeer returns a new Socketeer instance
// with a new DB and WebSocket instance.
//
// This method has to be exclusively called as per the requirements
// of the implementation and needs.
//
// # Parameters:
//
// 	- uriString (string): the MongoDB connection string.
// 	- dbName (string): the MongoDB database name.
// 	- collName (string): the MongoDB collection name.
//
// # Example:
//
// 	s, err := socketeer.NewSocketeer(uri, dbName, collName)
func NewSocketeer(uriString string, dbName string, collName string) (*Socketeer, error) {
	db, err := db.Connect(uriString, dbName, collName)
	if err != nil {
		return nil, err
	}

	return &Socketeer{
		DB: db,
		WS: ws.NewWebSocket(),
	}, nil
}

// Start starts the socketeer by starting the WebSocket server
// and listening for changes in the database.
//
// This method has to be exclusively called as per the requirements
// of the implementation and needs.
//
// # Parameters:
//
// 	- keys ([]string): the keys to listen for changes on.
// 	- host (string): the host address to listen on, example: localhost:8080
// 	- endpoint (string): the endpoint to listen on (without the trailing slash),
// 		example: /listen
//
// # Example:
//
// 	s.Start([]string{"title", "text"}, "localhost:8080", "/listen")
func (s *Socketeer) Start(keys []string, host string, endpoint string) error {
	fmt.Printf("Socketeer started\nVersion: %s", Version)

	go s.WS.Start(host, endpoint)

	err := s.DB.Listen(s.WS, keys)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// Stop stops the socketeer by stopping the WebSocket server
// and disconnecting from the database.
//
// This method has to be exclusively called as per the requirements
// of the implementation and needs.
//
// # Example:
//
// 	s.Stop()
func (s *Socketeer) Stop() error {
	defer func() {
		s.Stop()
		fmt.Println("Socketeer stopped gracefully.")
	}()

	s.DB.Disconnect()
	s.WS.Stop()

	return nil
}
