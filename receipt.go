package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"
)

type ReceiptScore struct {
	Id      uuid.UUID
	Receipt Receipt
	Points  int
}

type Receipt struct {
	Retailer     string `regex:"^[\\w\\s\\-&]+$"`
	PurchaseDate string `regex:"^\\d{4}-(0[1-9]|1[0-2])-([0-2]\\d|3[0-1])$"`
	PurchaseTime string `regex:"^([01]\\d|2[0-3]):([0-5]\\d)$"`
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
	points := calculatePoints(receipt)
	i.receipts[id] = ReceiptScore{Id: id, Receipt: receipt, Points: points}
	return nil
}

func calculatePoints(receipt Receipt) int {
	sum := 0
	sum += namePoints(receipt.Retailer)
	sum += roundDollarPoints(receipt.Total)
	sum += multiplesOfQuartersPoints(receipt.Total)
	sum += itemPairPoints(len(receipt.Items))
	sum += itemPoints(receipt.Items)
	sum += purchaseDatePoints(receipt.PurchaseDate)
	sum += purchaseTimePoints(receipt.PurchaseTime)
	return sum
}

func namePoints(name string) int {
	total := 0
	for _, char := range name {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			total += 1
		}
	}
	return total
}

func roundDollarPoints(total string) int {
	if total[len(total)-2:] == "00" {
		return 50
	}
	return 0
}

func multiplesOfQuartersPoints(total string) int {
	float, err := strconv.ParseFloat(total, 32)
	if err != nil {
		return 0
	}
	num := int(float * 100)
	if num%25 == 0 {
		return 25
	}
	return 0
}

func itemPairPoints(count int) int {
	return count / 2 * 5
}

func itemPoints(items []Item) int {
	total := 0
	for _, item := range items {
		total += itemDescriptionPoints(item)
	}
	return total
}

func itemDescriptionPoints(item Item) int {
	trimmedLength := len(strings.Trim(item.ShortDescription, " "))
	if trimmedLength%3 == 0 {
		itemPrice, err := strconv.ParseFloat(item.Price, 32)
		if err != nil {
			return 0
		}
		return int(math.Ceil(itemPrice * 0.2))
	} else {
		return 0
	}
}

func purchaseDatePoints(date string) int {
	dayString := date[len(date)-2:]
	day, err := strconv.Atoi(dayString)
	if err != nil {
		return 0
	}
	if day%2 == 0 {
		return 0
	}
	return 6
}

func purchaseTimePoints(time string) int {
	hour := time[:2]
	if hour == "14" || hour == "15" {
		minutes := time[3:]
		if hour == "14" && minutes == "00" {
			return 0
		} else {
			return 10
		}
	}
	return 0
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
