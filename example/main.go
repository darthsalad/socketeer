package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/darthsalad/socketeer"
	"github.com/joho/godotenv"
)

var (
	Version = "0.0.1"
	Build = "0"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB")
	collName := os.Getenv("MONGODB_COLLECTION")

	s, err := socketeer.NewSocketeer(uri, dbName, collName)
	if err != nil {
		log.Fatal(err)
	}

	defer func(){
		s.Stop()
		fmt.Println("Socketeer stopped gracefully.")
	}()

	fmt.Printf("Socketeer started\nVersion: %s Build: %s \n", Version, Build)

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-sigCh

	s.Stop()	
	fmt.Println("Socketeer stopped gracefully.")
}
