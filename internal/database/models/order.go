package models

import "github.com/uptrace/bun"

type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota
	OrderStatusConfirmed
	OrderStatusCanceled
)

func (o OrderStatus) String() string {
	return [...]string{"Pending", "Confirmed", "Delivering", "Completed", "Canceled"}[o-1]
}

func (o OrderStatus) EnumIndex() int {
	return int(o)
}

type Order struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID      int64 `bun:",pk,autoincrement"`
	OrderID string
	Price   float64

	// Notice how we avoid the M2M table making an string with the ID of the product
	ProductID string

	// Quantity of the product
	Quantity int64

	Status OrderStatus

	// Same for the user, we avoid the One-To-Many making an string with the ID of the user
	UserID int64
}
