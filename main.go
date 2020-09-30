package main

import (
	"fmt"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/server"
)

func main() {
	c := config.DefaultConfig()
	handler, err := server.NewHashEncoder(c)

	if err != nil {
		fmt.Printf("Creating the hash encoder service resulted in an error, see below \n %+v", err)
		return
	}

	if err := handler.Serve(); err != nil {
		fmt.Printf("Unexpected server shutdown error: %+v\n", err)
	} else {
		fmt.Println("Server exited properly")
	}
}
