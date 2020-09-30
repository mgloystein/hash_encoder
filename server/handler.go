package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/mgloystein/hash_encoder/common"
	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/service"
)

func NewHashEncoder(c *config.Config, shutdown chan os.Signal) (*HashEncoder, error) {
	svc, err := service.NewHasEncoderService(c)
	if err != nil {
		return nil, err
	}
	return &HashEncoder{
		service:  svc,
		shutdown: shutdown,
	}, nil
}

// Http handler for has encoding passwords
type HashEncoder struct {
	service  *service.HashEncoderService
	shutdown chan os.Signal
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
		itemID := h.service.CreateHash(passwd)
		res.Header().Add("location", fmt.Sprintf("/hash/%d", itemID))
		writeResult(res, http.StatusCreated, itemID)
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
		if item, err := h.service.GetHashedItem(int(itemID)); err != nil {
			writeError(res, http.StatusNotFound, err)
		} else {
			writeResult(res, http.StatusOK, item)
		}
	} else {
		writeError(res, http.StatusUnprocessableEntity, err)
	}
}

func (h *HashEncoder) Shutdown() {
	if h.shutdown != nil {
		h.shutdown <- syscall.SIGTERM
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
			stats := h.service.Stats()
			res.Header().Set("Content-Type", "application/json")
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

func notFoundResponse(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	writeResult(res, http.StatusNotFound, &common.MessageResult{"Can't find requested resource"})
}

func methodNotAllowedResponse(res http.ResponseWriter, method string) {
	res.Header().Set("Content-Type", "application/json")
	writeResult(res, http.StatusMethodNotAllowed, &common.MessageResult{fmt.Sprintf("%s is not allowed", method)})
}

func writeError(res http.ResponseWriter, status int, err error) {
	res.Header().Set("Content-Type", "application/json")
	http.Error(res, err.Error(), status)
}

func writeResult(res http.ResponseWriter, status int, response interface{}) {
	encoder := json.NewEncoder(res)
	res.WriteHeader(status)
	encoder.Encode(response)
}
