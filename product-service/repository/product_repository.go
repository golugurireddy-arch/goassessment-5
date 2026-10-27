package repository

import (
	"context"
	"errors"
	"fmt"

	"product-service/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("product not found")

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context) ([]model.Product, error)
}

type PostgresProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) (ProductRepository, error) {
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate products: %w", err)
	}
	return &PostgresProductRepository{db: db}, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var prod model.Product
	if err := r.db.WithContext(ctx).First(&prod, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &prod, nil
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]model.Product, error) {
	var prods []model.Product
	if err := r.db.WithContext(ctx).Find(&prods).Error; err != nil {
		return nil, err
	}
	return prods, nil
}
