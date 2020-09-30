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

	return &HashEncoderService{
		enigma: enigma,
		store:  store,
		timing: map[int]int64{},
		lock:   &lock,
		c:      c,
	}, nil
}

type HashEncoderService struct {
	enigma *hasher.Enigma
	store  storage.DataStore
	timing map[int]int64
	lock   *sync.RWMutex
	c      *config.Config
}

func (h *HashEncoderService) CreateHash(input string) int {
	itemPerstable := h.store.Reserve()
	go h.writeToStore(input, itemPerstable)
	return itemPerstable.ID()
}

func (h *HashEncoderService) GetHashedItem(id int) (string, error) {
	return h.store.Get(id)
}

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

func (h *HashEncoderService) writeToStore(input string, itemPerstable storage.Persistable) {
	time.Sleep(h.c.WriteDelay * time.Second)
	h.lock.Lock()
	start := time.Now().UnixNano()
	hashed, _ := h.enigma.Generate(input)
	h.timing[itemPerstable.ID()] = time.Now().UnixNano() - start
	h.lock.Unlock()
	if err := itemPerstable.Persist(hashed); err != nil {
		fmt.Printf("And error occured saving item %d, see below\n%+v\n", itemPerstable.ID(), err)
	}
}
