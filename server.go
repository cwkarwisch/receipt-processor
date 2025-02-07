package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

const jsonContentType = "application/json"

// used to encode the response to the POST /receipts/process route
type ID struct {
	Id uuid.UUID `json:"id"`
}

// used to encode the response to the GET /receipts/{id}/points
type Points struct {
	Points int `json:"points"`
}

type ReceiptServer struct {
	store ReceiptStore
	http.Handler
}

type ReceiptStore interface {
	GetPoints(uuid.UUID) (int, error)
	ProcessReceipt(uuid.UUID, io.Reader) error
}

func NewReceiptServer(store ReceiptStore) *ReceiptServer {
	router := http.NewServeMux()

	rs := &ReceiptServer{store: store}

	router.Handle("GET /receipts/{id}/points", http.HandlerFunc(rs.getReceiptPointsTotal))
	router.Handle("POST /receipts/process", http.HandlerFunc(rs.processReceipt))
	rs.Handler = router

	return rs
}

func (rs *ReceiptServer) getReceiptPointsTotal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	points, err := rs.store.GetPoints(uuid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", jsonContentType)
	err = json.NewEncoder(w).Encode(Points{points})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (rs *ReceiptServer) processReceipt(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	err := rs.store.ProcessReceipt(id, r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uuid := ID{id}
	json.NewEncoder(w).Encode(uuid)
	w.Header().Set("Content-Type", jsonContentType)
}
