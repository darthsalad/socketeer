package socketeer

import (
	// "fmt"
	"log"

	"github.com/darthsalad/socketeer/internal/db"
	// "github.com/darthsalad/socketeer/internal/ws"
)

type Socketeer struct {
	DB *db.DB
	// WS *ws.WS
}

func NewSocketeer(uriString string, dbName string, collName string) (*Socketeer, error) {
	db, err := db.Connect(uriString, dbName, collName)
	if err != nil {
		return nil, err
	}

	return &Socketeer{
		DB: db,
		// WS: ws,
	}, nil
}

func (s *Socketeer) Start() error {
	err := s.DB.Listen()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (s *Socketeer) Stop() error {
	err := s.DB.Disconnect()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}