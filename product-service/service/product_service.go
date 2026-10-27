package service

import (
	"context"
	"errors"
	"fmt"

	"product-service/model"
	"product-service/repository"
)

type ProductService interface {
	GetProduct(ctx context.Context, id int64) (*model.Product, error)
	ListProducts(ctx context.Context) ([]model.Product, error)
	CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*model.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.List(ctx)
}

func (s *productService) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than zero")
	}
	prod := &model.Product{
		Name:    req.Name,
		Price:   req.Price,
		InStock: *req.InStock,
	}
	if err := s.repo.Create(ctx, prod); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}
	return prod, nil
}
