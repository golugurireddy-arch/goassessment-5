package model

import "time"

type Product struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Price     float64   `gorm:"type:numeric(10,2);not null" json:"price"`
	InStock   bool      `gorm:"not null;default:true" json:"inStock"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateProductRequest struct {
	Name    string  `json:"name" binding:"required"`
	Price   float64 `json:"price" binding:"required,gt=0"`
	InStock *bool   `json:"inStock" binding:"required"`
}
