package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	badger "github.com/dgraph-io/badger/v4"
)

type BackendRequest struct {
	Resource string            `json:"resource"`
	Token    string            `json:"token"`
	Headers  map[string]string `json:"headers"`
}

type BackendResponse struct {
	DatabaseURL string `json:"database_url"`
}

func Serve(db *badger.DB) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		backendReq := BackendRequest{}

		if err := json.NewDecoder(req.Body).Decode(&backendReq); err != nil {
			rw.WriteHeader(400)
			fmt.Fprintf(rw, "error while parsing request: %v", err)
			return
		}

		resp := BackendResponse{}
		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(backendReq.Resource))
			if err != nil {
				rw.WriteHeader(404)
				return err
			}
			err = item.Value(func(val []byte) error {
				resp.DatabaseURL = string(val)
				return nil
			})
			if err != nil {
				log.Printf("API: error while reading value: %v", err)
				rw.WriteHeader(500)
				return err
			}
			return nil
		})
		if err != nil {
			return
		}

		_ = json.NewEncoder(rw).Encode(resp)
	})

	mux.HandleFunc("/list", func(rw http.ResponseWriter, req *http.Request) {
		var keys []string
		err := db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 10
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()
				keys = append(keys, string(k))
			}
			return nil
		})
		if err != nil {
			rw.WriteHeader(500)
			return
		}
		_ = json.NewEncoder(rw).Encode(keys)
	})

	srv := &http.Server{
		Addr:        ":4567",
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
