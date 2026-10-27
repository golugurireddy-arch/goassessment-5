package repository

import (
	"context"
	"errors"
	"fmt"

	"order-service/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	List(ctx context.Context) ([]*model.Order, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error)
}

type PostgresOrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) (OrderRepository, error) {
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate orders: %w", err)
	}
	return &PostgresOrderRepository{db: db}, nil
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("order %s not found", id)
		}
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepository) List(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, "id = ?", id).Error; err != nil {
			return err
		}
		if err := model.ValidateStatusTransition(order.Status, status); err != nil {
			return err
		}
		return tx.Model(&order).Update("status", status).Error
	})
	if err != nil {
		return nil, err
	}
	order.Status = status
	return &order, nil
}
