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
	db := database.NewMockDatabase(t, &models.Inventory{})
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
			expectedBody:   "Inventory service is running",
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

func TestGetListInventoryEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		endpointURL    string
		populateDB     func(db *bun.DB) []models.Inventory
	}{
		{
			name:           "GET request should return 204 OK and return empty inventory",
			method:         "GET",
			expectedStatus: http.StatusNoContent,
			endpointURL:    "/inventory",
			populateDB: func(db *bun.DB) []models.Inventory {
				return []models.Inventory{}
			},
		},
		{
			name:           "GET request should return 200 OK and return one inventory",
			method:         "GET",
			endpointURL:    "/inventory",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) []models.Inventory {
				inventory := &models.Inventory{
					ProductID: "1",
					Quantity:  1,
				}

				_, _ = db.NewInsert().Model(inventory).Returning("*").Exec(context.Background())

				return []models.Inventory{*inventory}
			},
		},
		{
			name:           "GET request should return 200 OK and return multiple inventory",
			method:         "GET",
			endpointURL:    "/inventory",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) []models.Inventory {
				inventoryList := []models.Inventory{}

				for i := 0; i < 10; i++ {
					inventory := &models.Inventory{
						ProductID: fmt.Sprintf("%d", i),
						Quantity:  int64(i),
					}

					inventoryList = append(inventoryList, *inventory)
				}

				_, _ = db.NewInsert().Model(&inventoryList).Returning("*").Exec(context.Background())

				return inventoryList
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

			if tt.expectedStatus == http.StatusNoContent {
				if len(expectedOrders) != 0 {
					t.Errorf("expected 0 inventory, got %d", len(expectedOrders))
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			var inventory []models.Inventory
			err = json.Unmarshal(body, &inventory)

			if err != nil {
				t.Fatalf("failed to parse JSON: %v; raw body: %s", err, string(body))
			}

			if len(inventory) != len(expectedOrders) {
				t.Errorf("expected %d inventory, got %d", len(expectedOrders), len(inventory))
			}

			for i, inventory := range inventory {
				if inventory.ID != expectedOrders[i].ID {
					t.Errorf("expected inventory %v, got %v", expectedOrders[i], inventory)
				}
			}
		})
	}
}

func TestGetUniqueInventoryEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		endpointURL    string
		populateDB     func(db *bun.DB) *models.Inventory
	}{
		{
			name:           "GET request should return 200 OK and return one inventory",
			method:         "GET",
			endpointURL:    "/inventory/1",
			expectedStatus: http.StatusOK,
			populateDB: func(db *bun.DB) *models.Inventory {
				inventory := &models.Inventory{
					ProductID: "1",
					Quantity:  1,
				}

				_, _ = db.NewInsert().Model(inventory).Returning("*").Exec(context.Background())

				return inventory
			},
		},
		{
			name:           "GET request should return 404 Not Found when inventory does not exist",
			method:         "GET",
			endpointURL:    "/inventory/100",
			expectedStatus: http.StatusNotFound,
			populateDB: func(db *bun.DB) *models.Inventory {
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

			expectedInventory := tt.populateDB(db)

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

			if expectedInventory == nil {
				return
			}

			var inventory *models.Inventory
			err = json.Unmarshal(body, &inventory)

			if err != nil {
				t.Fatalf("failed to parse JSON: %v; raw body: %s", err, string(body))
			}

			if inventory.ID != expectedInventory.ID {
				t.Errorf("expected inventory %v, got %v", expectedInventory, inventory)
			}
		})
	}
}
