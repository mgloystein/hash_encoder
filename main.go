package main

import (
	"fmt"
	"os"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/server"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		for _, arg := range args {
			if arg == "--help" {
				fmt.Println(`
hash-server [options]
	options:
	  -p   Sets the server port          (default 8080)
	  -w   Sets the service worker count (default 10)
	  -d   Sets the write delay          (default 5 seconds)
	  -S   Sets the hashing secret


hash-server usage:

    hash-server -p 8080 -w 4 -d 10
				`)
				return
			}
		}
	}

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
