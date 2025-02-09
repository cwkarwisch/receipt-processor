package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sync"

	"github.com/google/uuid"
)

type ReceiptScore struct {
	Id      uuid.UUID
	Receipt Receipt
	Points  int
}

type Receipt struct {
	Retailer     string `regex:"^[\\w\\s\\-&]+$"`
	PurchaseDate string
	PurchaseTime string
	Items        []Item
	Total        string `regex:"^\\d+\\.\\d{2}$"`
}

type Item struct {
	ShortDescription string `regex:"^[\\w\\s\\-]+$"`
	Price            string `regex:"^\\d+\\.\\d{2}$"`
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
		return err
	}

	err = validateStruct(receipt)

	if err != nil {
		return err
	}

	if len(receipt.Items) == 0 {
		return errors.New("no items included on receipt")
	}

	for _, item := range receipt.Items {
		err = validateStruct(item)

		if err != nil {
			return err
		}
	}

	i.receipts[id] = ReceiptScore{Id: id, Receipt: receipt, Points: 5}
	return nil
}

// used to validate the regex tags on Receipt and Item
func validateStruct[T any](data T) error {
	val := reflect.ValueOf(data)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		regexTag := field.Tag.Get("regex")
		if regexTag == "" {
			continue
		}

		regex, err := regexp.Compile(regexTag)
		if err != nil {
			return fmt.Errorf("invalid regex %q in field %q: %v", regexTag, field.Name, err)
		}

		fieldValue := val.Field(i).String()
		if !regex.MatchString(fieldValue) {
			return fmt.Errorf("field %q value %q does not meet regex %q", field.Name, fieldValue, regexTag)
		}
	}

	return nil
}
