package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"order-service/model"
	"order-service/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProductClient struct {
	mu       sync.RWMutex
	products map[string]*model.Product
}

func newMockProductClient() *mockProductClient {
	return &mockProductClient{
		products: map[string]*model.Product{
			"1": {
				ID:      1,
				Name:    "Wireless Mouse",
				Price:   29.99,
				InStock: true,
			},
			"out-of-stock-id": {
				ID:      2,
				Name:    "Mechanical Keyboard",
				Price:   89.99,
				InStock: false,
			},
		},
	}
}

func (m *mockProductClient) GetProduct(ctx context.Context, productID string) (*model.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.products[productID]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}
	return p, nil
}

type inMemoryOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func newInMemoryOrderRepo() *inMemoryOrderRepo {
	return &inMemoryOrderRepo{
		orders: make(map[string]*model.Order),
	}
}

func (r *inMemoryOrderRepo) Create(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return fmt.Errorf("order %s already exists", order.ID)
	}
	cp := *order
	r.orders[order.ID] = &cp
	return nil
}

func (r *inMemoryOrderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order %s not found", id)
	}
	cp := *order
	return &cp, nil
}

func (r *inMemoryOrderRepo) List(ctx context.Context) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*model.Order, 0, len(r.orders))
	for _, o := range r.orders {
		cp := *o
		list = append(list, &cp)
	}
	return list, nil
}

func (r *inMemoryOrderRepo) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order %s not found", id)
	}

	if err := model.ValidateStatusTransition(order.Status, status); err != nil {
		return nil, err
	}

	order.Status = status
	cp := *order
	return &cp, nil
}

func TestConcurrentOrders_100Requests(t *testing.T) {
	repo := newInMemoryOrderRepo()
	prodClient := newMockProductClient()
	orderSvc := service.NewOrderService(repo, prodClient)

	const totalWorkers = 100
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})

	var successCount int64
	var failCount int64

	createdOrderIDs := make([]string, totalWorkers)

	for i := 0; i < totalWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-startBarrier

			customerID := fmt.Sprintf("cust-%03d", idx)
			qty := (idx % 3) + 1

			order, err := orderSvc.CreateOrder(context.Background(), customerID, "1", qty)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}

			atomic.AddInt64(&successCount, 1)
			createdOrderIDs[idx] = order.ID
		}(i)
	}

	close(startBarrier)
	wg.Wait()

	assert.Equal(t, int64(totalWorkers), successCount)
	assert.Equal(t, int64(0), failCount)

	allOrders, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, allOrders, totalWorkers)

	uniqueIDs := make(map[string]struct{}, totalWorkers)
	for _, id := range createdOrderIDs {
		require.NotEmpty(t, id)
		_, seen := uniqueIDs[id]
		assert.False(t, seen, "Duplicate order ID detected: %s", id)
		uniqueIDs[id] = struct{}{}
	}
}

func TestConcurrentStatusUpdates(t *testing.T) {
	repo := newInMemoryOrderRepo()
	prodClient := newMockProductClient()
	orderSvc := service.NewOrderService(repo, prodClient)

	order, err := orderSvc.CreateOrder(context.Background(), "cust-test", "1", 2)
	require.NoError(t, err)

	const updateWorkers = 50
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})

	for i := 0; i < updateWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startBarrier
			_, _ = orderSvc.UpdateOrderStatus(context.Background(), order.ID, model.StatusConfirmed)
		}()
	}

	close(startBarrier)
	wg.Wait()

	updatedOrder, err := orderSvc.GetOrder(context.Background(), order.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusConfirmed, updatedOrder.Status)
}
