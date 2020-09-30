package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type StorageSettings struct {
	Host     string `json:"host",yaml:"host"`
	Port     int    `json:"port",yaml:"port"`
	Username string `json:"username",yaml:"username"`
	Password string `json:"password",yaml:"password"`
}

type Config struct {
	MasterSecret    string          `json:"masterSecret",yaml:"masterSecret"`
	Port            int             `json:"port",yaml:"port"`
	WorkerCount     int             `json:"workers",yaml:"workers"`
	StorageType     string          `json:"storage",yaml:"storage"`
	StorageSettings StorageSettings `json:"storageSettings",yaml:"storageSettings"`
	WriteDelay      time.Duration   `json:"delay",yaml:"delay`
}

func DefaultConfig() *Config {
	masterSecret := "imarealtivelylongandsomewhatsecuresecret"
	port := 8080
	delay := time.Duration(5)
	workers := 10
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p":
			i++
			if v, err := strconv.ParseInt(args[i], 10, 64); err == nil {
				port = int(v)
			} else {
				fmt.Println("Port supplied as an argument but was not recognized, using default")
			}
			break

		case "-w":
			i++
			if v, err := strconv.ParseInt(args[i], 10, 64); err == nil {
				workers = int(v)
			} else {
				fmt.Println("Worker count supplied as an argument but was not recognized, using default")
			}
			break

		case "-d":
			i++
			if v, err := strconv.ParseInt(args[i], 10, 64); err == nil {
				delay = time.Duration(v)
			} else {
				fmt.Println("Delay supplied as an argument but was not recognized, using default")
			}
			break

		case "-S":
			i++
			masterSecret = args[i]
			break

		}
	}

	return &Config{
		StorageType:  "memory",
		Port:         port,
		WorkerCount:  workers,
		MasterSecret: masterSecret,
		WriteDelay:   delay,
	}
}
