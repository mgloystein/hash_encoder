package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/hasher"
	"github.com/mgloystein/hash_encoder/storage"
)

func NewHashEncoder(c *config.Config, shutdown chan os.Signal) (*HashEncoder, error) {
	var lock sync.RWMutex
	enigma, err := hasher.NewEnigma(c.MasterSecret)

	if err != nil {
		return nil, fmt.Errorf(`Could not create Enigma, "%s"`, err.Error())
	}

	store, err := storage.NewDataStore(c)

	if err != nil {
		return nil, fmt.Errorf(`Could not create data storage, "%s"`, err.Error())
	}

	return &HashEncoder{
		enigma:   enigma,
		store:    store,
		timing:   map[int]int64{},
		lock:     &lock,
		c:        c,
		shutdown: shutdown,
	}, nil
}

// Http handler for has encoding passwords
type HashEncoder struct {
	enigma   *hasher.Enigma
	store    storage.DataStore
	timing   map[int]int64
	lock     *sync.RWMutex
	c        *config.Config
	shutdown chan os.Signal
}

func (h *HashEncoder) CreateHash(input string) int {
	itemPerstable := h.store.Reserve()
	go h.writeToStore(input, itemPerstable)
	return itemPerstable.ID()
}

func (h *HashEncoder) GetHashedItem(id int) (string, error) {
	return h.store.Get(id)
}

func (h *HashEncoder) Shutdown() {
	if h.shutdown != nil {
		h.shutdown <- syscall.SIGTERM
	}
}

func (h *HashEncoder) Stats() *stats {
	result := &stats{}
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

func (h *HashEncoder) HashPostRequest(res http.ResponseWriter, req *http.Request) {
	passwd := req.PostFormValue("password")

	if passwd == "" {
		if err := req.ParseForm(); err != nil {
			writeError(res, http.StatusUnprocessableEntity, err)
			return
		}
		passwd = req.Form.Get("password")
	}
	if passwd != "" {
		writeResult(res, http.StatusCreated, h.CreateHash(passwd))
		return
	} else {
		writeError(res, http.StatusUnprocessableEntity, fmt.Errorf("Required input: password"))
		return
	}
}

func (h *HashEncoder) HashGetRequest(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(strings.ToLower(req.URL.Path), "/")
	parts := strings.Split(path, "/")

	if itemID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		if item, err := h.GetHashedItem(int(itemID)); err != nil {
			writeError(res, http.StatusNotFound, err)
		} else {
			writeResult(res, http.StatusOK, item)
		}
	} else {
		writeError(res, http.StatusUnprocessableEntity, err)
	}
}

func (h *HashEncoder) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(strings.ToLower(req.URL.Path), "/")
	parts := strings.Split(path, "/")
	base := parts[0]
	fmt.Printf("Recording access to %s %s\n", req.Method, path)

	switch base {
	case "hash":
		if len(parts) == 2 && req.Method == "GET" {
			h.HashGetRequest(res, req)
			return
		} else if len(parts) == 1 && req.Method == "POST" {
			h.HashPostRequest(res, req)
			return
		}

		break
	case "stats":
		if req.Method == "GET" {
			stats := h.Stats()
			writeResult(res, http.StatusOK, stats)
			return
		}
		break
	case "shutdown":
		res.WriteHeader(http.StatusNoContent)
		go h.Shutdown()
		return
	default:
		notFoundResponse(res)
		return
	}
	methodNotAllowedResponse(res, req.Method)
}

func (h *HashEncoder) writeToStore(input string, itemPerstable storage.Persistable) {
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

func notFoundResponse(res http.ResponseWriter) {
	writeResult(res, http.StatusNotFound, &messageResult{"Can't find requested resource"})
}

func methodNotAllowedResponse(res http.ResponseWriter, method string) {
	writeResult(res, http.StatusMethodNotAllowed, &messageResult{fmt.Sprintf("%s is not allowed", method)})
}

func writeError(res http.ResponseWriter, status int, err error) {
	http.Error(res, err.Error(), status)
}

func writeResult(res http.ResponseWriter, status int, response interface{}) {
	encoder := json.NewEncoder(res)
	res.WriteHeader(status)
	encoder.Encode(response)
}
