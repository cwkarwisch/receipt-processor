package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestProcessingReceiptsAndFetchingThem(t *testing.T) {
	targetReceipt := `{
		"retailer": "Target",
		"purchaseDate": "2022-01-01",
		"purchaseTime": "13:01",
		"items": [
			{
			"shortDescription": "Mountain Dew 12PK",
			"price": "6.49"
			},{
			"shortDescription": "Emils Cheese Pizza",
			"price": "12.25"
			},{
			"shortDescription": "Knorr Creamy Chicken",
			"price": "1.26"
			},{
			"shortDescription": "Doritos Nacho Cheese",
			"price": "3.35"
			},{
			"shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
			"price": "12.00"
			}
		],
		"total": "35.35"
		}`

	cornerMarketReceipt := `{
		"retailer": "M&M Corner Market",
		"purchaseDate": "2022-03-20",
		"purchaseTime": "14:33",
		"items": [
			{
			"shortDescription": "Gatorade",
			"price": "2.25"
			},{
			"shortDescription": "Gatorade",
			"price": "2.25"
			},{
			"shortDescription": "Gatorade",
			"price": "2.25"
			},{
			"shortDescription": "Gatorade",
			"price": "2.25"
			}
		],
		"total": "9.00"
		}`

	store := NewReceiptStore()
	server := NewReceiptServer(store)

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newPostReceiptRequest(targetReceipt))
	assertResponseCode(t, response.Code, http.StatusOK)
	var firstId ID
	err := json.NewDecoder(response.Body).Decode(&firstId)
	checkDecodeErr(t, response, err)

	response = httptest.NewRecorder()
	server.ServeHTTP(response, newPostReceiptRequest(cornerMarketReceipt))
	assertResponseCode(t, response.Code, http.StatusOK)
	var secondId ID
	err = json.NewDecoder(response.Body).Decode(&secondId)
	checkDecodeErr(t, response, err)

	response = httptest.NewRecorder()
	server.ServeHTTP(response, newGetPointsRequest(firstId.Id))
	assertResponseCode(t, response.Code, http.StatusOK)
	assertPointTotalInResponse(t, response, 28)

	response = httptest.NewRecorder()
	server.ServeHTTP(response, newGetPointsRequest(secondId.Id))
	assertResponseCode(t, response.Code, http.StatusOK)
	assertPointTotalInResponse(t, response, 109)
}

func TestGetReceiptPoints(t *testing.T) {
	store := NewReceiptStore()
	id := uuid.New()
	store.receipts[id] = ReceiptScore{Id: id, Receipt: Receipt{}, Points: 100}
	server := NewReceiptServer(store)

	t.Run("returns receipt points", func(t *testing.T) {
		request := newGetPointsRequest(id)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertContentType(t, response.Header(), "application/json")
		assertPointTotalInResponse(t, response, 100)
	})

	t.Run("request is made with id for non-existent receipt", func(t *testing.T) {
		request := newGetPointsRequest(uuid.New())
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, response.Body.String(), notFoundMessage+"\n")
	})
}

func TestProcessReceipt(t *testing.T) {
	store := NewReceiptStore()
	server := NewReceiptServer(store)
	receiptJson := `{
    "retailer": "Walgreens",
    "purchaseDate": "2022-01-02",
    "purchaseTime": "08:13",
    "total": "2.65",
    "items": [
        {"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
        {"shortDescription": "Dasani", "price": "1.40"}
    ]
}`

	t.Run("accepts POST of a new receipt", func(t *testing.T) {
		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertContentType(t, response.Header(), "application/json")

		var id ID
		err := json.NewDecoder(response.Body).Decode(&id)
		checkDecodeErr(t, response, err)
		savedReceipt, ok := store.receipts[id.Id]
		if !ok {
			t.Errorf("receipt was not saved during POST request")
		}
		if savedReceipt.Id != id.Id {
			t.Errorf("ReceiptScore did not properly record receipt ID")
		}
	})

	t.Run("rejects receipt with missing retailer", func(t *testing.T) {
		receiptJson = `{
			"retailer": "",
			"purchaseDate": "2022-05-27",
			"purchaseTime": "02:00",
			"total": "3.29",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with missing items", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2022-11-11",
			"total": "3.29"
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid formatting for a field", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2022-02-28",
			"purchaseTime": "08:13",
			"total": "3",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with missing short description", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2022-09-20",
			"purchaseTime": "12:01",
			"total": "3.29",
			"items": [
				{"shortDescription": "", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with missing item price", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2022-12-31",
			"purchaseTime": "23:50",
			"total": "3.29",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": ""}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid date format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "202-01-02",
			"purchaseTime": "08:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid month in date format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2000-13-02",
			"purchaseTime": "08:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid day in date format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2022-03-50",
			"purchaseTime": "08:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with 00 month in date format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2000-00-02",
			"purchaseTime": "08:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid day field in date format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "1024-10-32",
			"purchaseTime": "08:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with invalid time format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2021-01-02",
			"purchaseTime": "0:13",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with a different invalid time format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2021-01-02",
			"purchaseTime": "12:60",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})

	t.Run("rejects receipt with a final invalid time format", func(t *testing.T) {
		receiptJson = `{
			"retailer": "Walgreens",
			"purchaseDate": "2021-01-02",
			"purchaseTime": "77:77",
			"total": "3.15",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`

		request := newPostReceiptRequest(receiptJson)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, response.Body.String(), badRequestMessage+"\n")
	})
}

func newPostReceiptRequest(receipt string) *http.Request {
	body := []byte(receipt)
	req, _ := http.NewRequest(http.MethodPost, "/receipts/process", bytes.NewReader(body))
	return req
}

func newGetPointsRequest(id uuid.UUID) *http.Request {
	path := "/receipts/" + id.String() + "/points"
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return req
}

func assertResponseBody(t testing.TB, body, expected string) {
	t.Helper()
	if body != expected {
		t.Errorf("expected response body of %q but got %q", expected, body)
	}
}

func assertPointTotalInResponse(t testing.TB, response *httptest.ResponseRecorder, expected int) {
	t.Helper()
	var points Points
	err := json.NewDecoder(response.Body).Decode(&points)
	checkDecodeErr(t, response, err)
	if points.Points != expected {
		t.Errorf("expected points total of %d but got %d", expected, points.Points)
	}
}

func assertContentType(t testing.TB, header http.Header, want string) {
	t.Helper()
	got := header.Get("Content-Type")

	if got != want {
		t.Errorf("expected Content-Type Header of %q but got %q", want, got)
	}
}

func assertResponseCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expect response code of %d but got %d", want, got)
	}
}

func checkDecodeErr(t testing.TB, response *httptest.ResponseRecorder, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unable to parse response into ID. resp: %q, err: %v", response.Body, err)
	}
}
