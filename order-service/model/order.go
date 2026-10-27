package model

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusConfirmed, StatusCancelled:
		return true
	default:
		return false
	}
}

func ValidateStatusTransition(current, next OrderStatus) error {
	if current == StatusCancelled {
		return fmt.Errorf("cannot alter a CANCELLED order")
	}
	if current == StatusConfirmed && next == StatusCreated {
		return fmt.Errorf("cannot revert CONFIRMED order back to CREATED")
	}
	return nil
}

type Order struct {
	ID          string      `gorm:"primaryKey;size:64" json:"id"`
	CustomerID  string      `gorm:"not null;index" json:"customerId"`
	ProductID   string      `gorm:"not null;index" json:"productId"`
	Quantity    int         `gorm:"not null" json:"quantity"`
	TotalAmount float64     `gorm:"type:numeric(12,2);not null" json:"totalAmount"`
	Status      OrderStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CreateOrderDTO struct {
	CustomerID string `json:"customerId"`
	ProductID  string `json:"productId" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
}

type UpdateStatusDTO struct {
	Status OrderStatus `json:"status" binding:"required"`
}

type Product struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	InStock bool    `json:"inStock"`
}
