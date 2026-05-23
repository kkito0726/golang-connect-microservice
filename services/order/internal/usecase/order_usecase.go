package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ken/connect-microservice/services/order/internal/domain"
)

type OrderUsecase struct {
	repo          domain.OrderRepository
	productClient domain.ProductClient
	userClient    domain.UserClient
}

func NewOrderUsecase(repo domain.OrderRepository, productClient domain.ProductClient, userClient domain.UserClient) *OrderUsecase {
	return &OrderUsecase{repo: repo, productClient: productClient, userClient: userClient}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, input domain.CreateOrderInput) (domain.Order, error) {
	if err := uc.userClient.ValidateUser(ctx, input.UserID); err != nil {
		return domain.Order{}, fmt.Errorf("user not found: %w", err)
	}

	products := make(map[string]domain.ProductInfo, len(input.Items))
	for _, item := range input.Items {
		product, err := uc.productClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return domain.Order{}, fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}
		products[item.ProductID] = product
	}

	order, err := domain.NewOrder(input.UserID, input.Items, products)
	if err != nil {
		return domain.Order{}, err
	}

	var deducted []domain.OrderItem
	for _, item := range order.Items {
		if err := uc.productClient.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
			for _, d := range deducted {
				if rerr := uc.productClient.RestoreStock(ctx, d.ProductID, d.Quantity, ""); rerr != nil {
					slog.Error("saga compensation failed: stock restore",
						"product_id", d.ProductID, "error", rerr)
				}
			}
			return domain.Order{}, fmt.Errorf("deduct stock for %s: %w", item.ProductID, err)
		}
		deducted = append(deducted, item)
	}

	return uc.repo.Create(ctx, order)
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, userID, status string, pageSize, page int) ([]domain.Order, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, userID, status, pageSize, (page-1)*pageSize)
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, id, status string) (domain.Order, error) {
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, id string) (domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}

	cancelled, err := order.Cancel()
	if err != nil {
		return domain.Order{}, err
	}

	for _, item := range cancelled.Items {
		if err := uc.productClient.RestoreStock(ctx, item.ProductID, item.Quantity, order.ID); err != nil {
			return domain.Order{}, fmt.Errorf("restore stock for %s: %w", item.ProductID, err)
		}
	}
	return uc.repo.UpdateStatus(ctx, id, cancelled.Status)
}
