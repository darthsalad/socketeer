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

	fields := []string{"title", "text"}
	url := "localhost:8080"
	endpoint := "/listen"

	s.Start(fields, url, endpoint)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-sigCh

	s.Stop()	
	fmt.Println("Socketeer stopped gracefully.")
}
