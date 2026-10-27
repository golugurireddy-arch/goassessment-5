package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"order-service/client"
	"order-service/model"
	"order-service/repository"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("order not found")
	ErrOutOfStock = errors.New("product is out of stock")
)

type OrderService interface {
	CreateOrder(ctx context.Context, customerID, productID string, quantity int) (*model.Order, error)
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	ListOrders(ctx context.Context) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error)
}

type orderService struct {
	repo          repository.OrderRepository
	productClient client.ProductClient
}

func NewOrderService(repo repository.OrderRepository, pc client.ProductClient) OrderService {
	return &orderService{repo: repo, productClient: pc}
}

func (s *orderService) CreateOrder(ctx context.Context, customerID, productID string, quantity int) (*model.Order, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	product, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("external product check failed: %w", err)
	}

	if !product.InStock {
		return nil, ErrOutOfStock
	}

	now := time.Now().UTC()
	order := &model.Order{
		ID:          uuid.New().String(),
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		TotalAmount: float64(quantity) * product.Price,
		Status:      model.StatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return order, nil
}

func (s *orderService) ListOrders(ctx context.Context) ([]*model.Order, error) {
	return s.repo.List(ctx)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid order status: %s", status)
	}
	order, err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, ErrNotFound
	}
	return order, nil
}
