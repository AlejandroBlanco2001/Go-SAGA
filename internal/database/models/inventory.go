package models

import "github.com/uptrace/bun"

type Inventory struct {
	bun.BaseModel `bun:"table:inventory,alias:i"`

	ID        int64 `bun:",pk,autoincrement"`
	ProductID string
	Quantity  int64
}
