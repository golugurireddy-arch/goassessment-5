package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"order-service/model"
	"order-service/service"

	"github.com/stretchr/testify/assert"
)

type ConcurrentMockRepo struct {
	mu     sync.Mutex
	orders map[string]*model.Order
}

func (r *ConcurrentMockRepo) Create(ctx context.Context, o *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID] = o
	return nil
}

func (r *ConcurrentMockRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("not found")
}

func (r *ConcurrentMockRepo) List(ctx context.Context) ([]*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := make([]*model.Order, 0, len(r.orders))
	for _, v := range r.orders {
		list = append(list, v)
	}
	return list, nil
}

func (r *ConcurrentMockRepo) UpdateStatus(ctx context.Context, id string, s model.OrderStatus) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.orders[id]; ok {
		o.Status = s
		return o, nil
	}
	return nil, fmt.Errorf("not found")
}

type StubProductClient struct{}

func (s *StubProductClient) GetProduct(ctx context.Context, productID string) (*model.Product, error) {
	return &model.Product{
		ID:      501,
		Name:    "Load Test Item",
		Price:   15.50,
		InStock: true,
	}, nil
}

func TestConcurrentOrders_100Requests(t *testing.T) {
	repo := &ConcurrentMockRepo{orders: make(map[string]*model.Order)}
	prodClient := &StubProductClient{}
	svc := service.NewOrderService(repo, prodClient)

	concurrentRequests := 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	errChan := make(chan error, concurrentRequests)

	start := time.Now()
	for i := 0; i < concurrentRequests; i++ {
		custID := fmt.Sprintf("cust-%d", i)
		go func(cid string) {
			defer wg.Done()
			_, err := svc.CreateOrder(context.Background(), cid, "501", 1)
			if err != nil {
				errChan <- err
			}
		}(custID)
	}

	wg.Wait()
	close(errChan)

	duration := time.Since(start)
	t.Logf("Processed %d concurrent orders in %v", concurrentRequests, duration)

	for err := range errChan {
		assert.NoError(t, err)
	}

	orders, _ := repo.List(context.Background())
	assert.Equal(t, concurrentRequests, len(orders))
}
