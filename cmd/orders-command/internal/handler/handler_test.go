package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"saga-pattern/internal/database"
	"saga-pattern/internal/database/models"
)

func setupHandler(t *testing.T) (http.Handler, *bun.DB) {
	db := database.NewMockDatabase(t, &models.Order{})
	logger, _ := zap.NewDevelopment()
	handler := NewHandler(logger, db, context.Background(), nil)
	return handler, db
}

func TestHealthEndpoint(t *testing.T) {
	handler, _ := setupHandler(t)
	server := httptest.NewServer(handler)
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request should return 200 OK",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, server.URL+"/health", nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}
		})
	}
}

func TestGetListOrdersEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		endpointURL    string
		populateDB     func(db *bun.DB) []models.Order
	}{
		{
			name:           "GET request should return 200 OK and return empty orders",
			method:         "GET",
			expectedStatus: http.StatusOK,
			endpointURL:    "/orders",
			populateDB: func(db *bun.DB) []models.Order {
				return []models.Order{}
			},
		},
		{
			name:           "GET request should return 200 OK and return one order",
			method:         "GET",
			endpointURL:    "/orders",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) []models.Order {
				order := &models.Order{
					Price:     100,
					OrderID:   "1",
					ProductID: "1",
					Quantity:  1,
					UserID:    1,
				}

				_, _ = db.NewInsert().Model(order).Returning("*").Exec(context.Background())

				return []models.Order{*order}
			},
		},
		{
			name:           "GET request should return 200 OK and return multiple orders",
			method:         "GET",
			endpointURL:    "/orders",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) []models.Order {
				orders := []models.Order{}

				for i := 0; i < 10; i++ {
					order := &models.Order{
						Price:     float64(i),
						OrderID:   fmt.Sprintf("%d", i),
						ProductID: fmt.Sprintf("%d", i),
						Quantity:  int64(i),
						UserID:    int64(i),
					}

					orders = append(orders, *order)
				}

				_, _ = db.NewInsert().Model(&orders).Returning("*").Exec(context.Background())

				return orders
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, db := setupHandler(t)
			server := httptest.NewServer(handler)
			defer server.Close()

			url := fmt.Sprintf("%s%s", server.URL, tt.endpointURL)

			req, err := http.NewRequest(tt.method, url, nil)

			if err != nil {
				t.Fatal(err)
			}

			expectedOrders := tt.populateDB(db)

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatal(err)
			}

			var orders []models.Order
			err = json.Unmarshal(body, &orders)

			if err != nil {
				t.Fatalf("failed to parse JSON: %v; raw body: %s", err, string(body))
			}

			if len(orders) != len(expectedOrders) {
				t.Errorf("expected %d orders, got %d", len(expectedOrders), len(orders))
			}

			for i, order := range orders {
				if order.ID != expectedOrders[i].ID {
					t.Errorf("expected order %v, got %v", expectedOrders[i], order)
				}
			}
		})
	}
}

func TestGetUniqueOrdersEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		endpointURL    string
		populateDB     func(db *bun.DB) *models.Order
	}{
		{
			name:           "GET request should return 200 OK and return one order",
			method:         "GET",
			endpointURL:    "/orders/1",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) *models.Order {
				order := &models.Order{
					Price:     100,
					OrderID:   "1",
					ProductID: "1",
					Quantity:  1,
					UserID:    1,
				}

				_, _ = db.NewInsert().Model(order).Returning("*").Exec(context.Background())

				return order
			},
		},
		{
			name:           "GET request should return 404 Not Found when order does not exist",
			method:         "GET",
			endpointURL:    "/orders/100",
			expectedStatus: http.StatusNotFound,
			populateDB: func(db *bun.DB) *models.Order {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, db := setupHandler(t)
			server := httptest.NewServer(handler)
			defer server.Close()

			url := fmt.Sprintf("%s%s", server.URL, tt.endpointURL)

			req, err := http.NewRequest(tt.method, url, nil)

			if err != nil {
				t.Fatal(err)
			}

			expectedOrder := tt.populateDB(db)

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatal(err)
			}

			if expectedOrder == nil {
				return
			}

			var order *models.Order
			err = json.Unmarshal(body, &order)

			if err != nil {
				t.Fatalf("failed to parse JSON: %v; raw body: %s", err, string(body))
			}

			if order.ID != expectedOrder.ID {
				t.Errorf("expected order %v, got %v", expectedOrder, order)
			}
		})
	}
}
