package main

import (
	"fmt"
	"os"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/command"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/dict"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/server"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/store"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	d := dict.New()
	s := store.New(d)
	c := command.New(s)

	port := os.Getenv("REDIS_PORT")
	server := server.New(server.Config{Port: port}, c.Run)

	if err := server.Run(); err != nil {
		fmt.Println("Error starting Redis server: ", err.Error())
		os.Exit(1)
	}
}
