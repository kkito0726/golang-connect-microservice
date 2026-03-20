package usecase

import (
	"context"
	"fmt"

	"github.com/ken/connect-microservice/services/order/internal/repository"
)

// ProductInfo はユースケースが外部サービスから必要とする商品情報。
// proto 型への依存を避けるためここで定義する。
type ProductInfo struct {
	ID            string
	Name          string
	PriceCents    int64
	StockQuantity int32
}

// UserClient はユーザーサービスへの依存を抽象化するインターフェース。
type UserClient interface {
	ValidateUser(ctx context.Context, id string) error
}

// ProductClient は商品サービスへの依存を抽象化するインターフェース。
type ProductClient interface {
	GetProduct(ctx context.Context, id string) (ProductInfo, error)
	DeductStock(ctx context.Context, productID string, quantity int32) error
	RestoreStock(ctx context.Context, productID string, quantity int32, referenceID string) error
}

// OrderRepository はデータ永続化への依存を抽象化するインターフェース。
type OrderRepository interface {
	Create(ctx context.Context, userID string, items []repository.OrderItem, totalCents int64) (repository.Order, error)
	GetByID(ctx context.Context, id string) (repository.Order, error)
	List(ctx context.Context, userID, status string, limit, offset int) ([]repository.Order, int, error)
	UpdateStatus(ctx context.Context, id, status string) (repository.Order, error)
}

type CreateOrderInput struct {
	UserID string
	Items  []CreateOrderItem
}

type CreateOrderItem struct {
	ProductID string
	Quantity  int32
}

type OrderUsecase struct {
	repo          OrderRepository
	productClient ProductClient
	userClient    UserClient
}

func NewOrderUsecase(repo OrderRepository, productClient ProductClient, userClient UserClient) *OrderUsecase {
	return &OrderUsecase{repo: repo, productClient: productClient, userClient: userClient}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (repository.Order, error) {
	if err := uc.userClient.ValidateUser(ctx, input.UserID); err != nil {
		return repository.Order{}, fmt.Errorf("user not found: %w", err)
	}

	var orderItems []repository.OrderItem
	var totalCents int64

	for _, item := range input.Items {
		product, err := uc.productClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return repository.Order{}, fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}
		if product.StockQuantity < item.Quantity {
			return repository.Order{}, fmt.Errorf("insufficient stock for %s: have %d, want %d",
				product.Name, product.StockQuantity, item.Quantity)
		}
		orderItems = append(orderItems, repository.OrderItem{
			ProductID:      item.ProductID,
			ProductName:    product.Name,
			Quantity:       item.Quantity,
			UnitPriceCents: product.PriceCents,
		})
		totalCents += product.PriceCents * int64(item.Quantity)
	}

	for _, item := range orderItems {
		if err := uc.productClient.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
			return repository.Order{}, fmt.Errorf("deduct stock for %s: %w", item.ProductID, err)
		}
	}

	return uc.repo.Create(ctx, input.UserID, orderItems, totalCents)
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, id string) (repository.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, userID, status string, pageSize, page int) ([]repository.Order, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, userID, status, pageSize, (page-1)*pageSize)
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, id, status string) (repository.Order, error) {
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, id string) (repository.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return repository.Order{}, err
	}
	if order.Status != "pending" {
		return repository.Order{}, fmt.Errorf("can only cancel pending orders, current status: %s", order.Status)
	}
	for _, item := range order.Items {
		if err := uc.productClient.RestoreStock(ctx, item.ProductID, item.Quantity, order.ID); err != nil {
			return repository.Order{}, fmt.Errorf("restore stock for %s: %w", item.ProductID, err)
		}
	}
	return uc.repo.UpdateStatus(ctx, id, "cancelled")
}
