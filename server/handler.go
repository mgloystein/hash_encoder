package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mgloystein/hash_encoder/common"
	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/service"
)

// NewHashEncoder returns a new HashEncoder HTTP handler
func NewHashEncoder(c *config.Config) (*HashEncoder, error) {
	svc, err := service.NewHasEncoderService(c)
	if err != nil {
		return nil, err
	}
	shutdown := make(chan os.Signal, 1)
	return &HashEncoder{
		service:  svc,
		shutdown: shutdown,
		c:        c,
	}, nil
}

// HashEncoder is an HTTP handler for hashing and encoding passwords
type HashEncoder struct {
	service  *service.HashEncoderService
	shutdown chan os.Signal
	c        *config.Config
}

// HashPostRequest handles POST requests to /hash
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

// HashGetRequest handles GET requests to /hash/:id
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

// Shutdown requests the server to terminate
func (h *HashEncoder) Shutdown() {
	if h.shutdown != nil {
		h.shutdown <- syscall.SIGTERM
	}
}

// Serve starts the HTTP and system singal listeners
func (h *HashEncoder) Serve() error {
	signal.Notify(h.shutdown, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", h.c.Port),
		Handler: h,
	}

	go func() {
		fmt.Printf("Starting server on port %d\n", h.c.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				fmt.Println("Server connection closed")
			default:
				fmt.Printf("Unexpected server error: %+v\n", err)
				h.shutdown <- syscall.SIGKILL
			}
		}
	}()

	ctx, cancel := handleSystemSignals(h.shutdown)

	defer func() {
		h.service.Terminate()
		if cancel != nil {
			cancel()
		}
	}()

	return httpServer.Shutdown(ctx)
}

// ServeHTTP is the initial responder to all HTTP requests
// TODO: Replace with Gin or Gorilla
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
	writeResult(res, http.StatusNotFound, &common.MessageResult{Message: "Can't find requested resource"})
}

func methodNotAllowedResponse(res http.ResponseWriter, method string) {
	res.Header().Set("Content-Type", "application/json")
	writeResult(res, http.StatusMethodNotAllowed, &common.MessageResult{Message: fmt.Sprintf("%s is not allowed", method)})
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

func handleSystemSignals(shutdown chan os.Signal) (context.Context, context.CancelFunc) {
	switch <-shutdown {
	case os.Interrupt:
	case syscall.SIGINT:
	case syscall.SIGTERM:
		return context.WithTimeout(context.Background(), 10*time.Second)
	default:
		return context.WithTimeout(context.Background(), 0*time.Second)
	}
	return nil, nil
}
