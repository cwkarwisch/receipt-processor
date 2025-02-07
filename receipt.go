package main

import (
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/google/uuid"
)

type ReceiptScore struct {
	Id      uuid.UUID
	Receipt Receipt
	Points  int
}

type Receipt struct {
	Retailer     string
	PurchaseDate string
	PurchaseTime string
	Items        []Item
	Total        string
}

type Item struct {
	ShortDescription string
	Price            string
}

type InMemoryReceiptStore struct {
	receipts map[uuid.UUID]ReceiptScore
	mu       sync.Mutex
}

func NewReceiptStore() *InMemoryReceiptStore {
	receipts := make(map[uuid.UUID]ReceiptScore)
	return &InMemoryReceiptStore{receipts: receipts}
}

func (i *InMemoryReceiptStore) GetPoints(id uuid.UUID) (int, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	receiptScore, ok := i.receipts[id]
	if ok {
		return receiptScore.Points, nil
	} else {
		return 0, errors.New("no receipt found")
	}
}

func (i *InMemoryReceiptStore) ProcessReceipt(id uuid.UUID, body io.Reader) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	var receipt Receipt
	err := json.NewDecoder(body).Decode(&receipt)

	if err != nil {
		return errors.New("could not decode request body")
	}
	i.receipts[id] = ReceiptScore{Id: id, Receipt: receipt, Points: 5}
	return nil
}
