package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/mgloystein/hash_encoder/common"
	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/hasher"
	"github.com/mgloystein/hash_encoder/storage"
)

// NewHasEncoderService creates a new service for hashing, encoding, and storing passwords
func NewHasEncoderService(c *config.Config) (*HashEncoderService, error) {
	var lock sync.RWMutex
	enigma, err := hasher.NewEnigma(c.MasterSecret)

	if err != nil {
		return nil, fmt.Errorf(`Could not create Enigma, "%s"`, err.Error())
	}

	store, err := storage.NewDataStore(c)

	if err != nil {
		return nil, fmt.Errorf(`Could not create data storage, "%s"`, err.Error())
	}

	service := &HashEncoderService{
		enigma:   enigma,
		store:    store,
		timing:   map[int]int64{},
		lock:     &lock,
		c:        c,
		queue:    make(chan *persistablePackage, c.WorkerCount),
		complete: make(chan bool, 1),
	}
	service.init()
	return service, nil
}

type persistablePackage struct {
	storage.Persistable
	value string
}

// HashEncoderService is a service to hash, encode, and store password passwords
type HashEncoderService struct {
	enigma   *hasher.Enigma
	store    storage.DataStore
	timing   map[int]int64
	lock     *sync.RWMutex
	c        *config.Config
	queue    chan *persistablePackage
	complete chan bool
}

func (h *HashEncoderService) init() {
	for i := 0; i < h.c.WorkerCount; i++ {
		fmt.Printf("Starting worker process: %d\n", i)
		go h.worker(i)
	}
}

func (h *HashEncoderService) worker(id int) {
	for true {
		pack := <-h.queue

		if pack == nil {
			fmt.Printf("Stopping worker process: %d\n", id)
			break
		}

		time.Sleep(h.c.WriteDelay * time.Second)
		itemPerstable := pack.Persistable

		hashed := h.generateHash(pack.value, itemPerstable.ID())
		itemPerstable.Value(hashed)

		h.writeToStore(itemPerstable)
	}
	h.complete <- true
}

// Terminate shuts down the worker queues
func (h *HashEncoderService) Terminate() {
	for i := 0; i < h.c.WorkerCount; i++ {
		h.queue <- nil
	}
	for i := 0; i < h.c.WorkerCount; i++ {
		<-h.complete
	}
}

// CreateHash reserves an ID for a hashed password, then send it to a woerk queue for processing
func (h *HashEncoderService) CreateHash(input string) int {
	itemPerstable := h.store.Reserve()
	h.queue <- &persistablePackage{itemPerstable, input}
	return itemPerstable.ID()
}

// GetHashedItem gets the hashed and encoded password by ID
func (h *HashEncoderService) GetHashedItem(id int) (string, error) {
	return h.store.Get(id)
}

// Stats allows introspection into the number of stored passwords and how long it took to process each
func (h *HashEncoderService) Stats() *common.Stats {
	result := &common.Stats{}
	h.lock.RLock()
	var total int64
	for _, timing := range h.timing {
		total += timing
		result.Count++
	}
	h.lock.RUnlock()
	if result.Count > 0 {
		result.AverageProcessTime = (float64(total) / float64(result.Count)) / 1000
	}
	return result
}

func (h *HashEncoderService) generateHash(value string, id int) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	start := time.Now().UnixNano()

	hashed, _ := h.enigma.Generate(value)

	h.timing[id] = time.Now().UnixNano() - start

	return hashed
}

func (h *HashEncoderService) writeToStore(itemPerstable storage.Persistable) {
	if err := itemPerstable.Persist(); err != nil {
		fmt.Printf("And error occured saving item %d, see below\n%+v\n", itemPerstable.ID(), err)
	}
	fmt.Printf("Successfully writen item %d\n", itemPerstable.ID())
}
