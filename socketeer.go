package socketeer

import (
	"fmt"
	"log"

	"github.com/darthsalad/socketeer/internal/db"
	"github.com/darthsalad/socketeer/internal/ws"
)

type Socketeer struct {
	DB *db.DB
	WS *ws.WebSocket
}

var (
	Version = "0.1.0"
	Build = "0"
)

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

func (s *Socketeer) Start() error {
	fmt.Printf("Socketeer started\nVersion: %s Build: %s \n", Version, Build)

	go s.WS.Start()
	
	err := s.DB.Listen(s.WS)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (s *Socketeer) Stop() error {
	defer func(){
		s.Stop()
		fmt.Println("Socketeer stopped gracefully.")
	}()

	s.DB.Disconnect()
	s.WS.Stop()
	
	return nil
}