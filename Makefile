.PHONEY: test build

test:
	go test github.com/mgloystein/hash_encoder/hasher
	go test github.com/mgloystein/hash_encoder/service
	go test github.com/mgloystein/hash_encoder/storage

build:
	go build -o hash-server .
