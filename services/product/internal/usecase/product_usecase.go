package usecase

import (
	"context"

	"github.com/ken/connect-microservice/services/product/internal/repository"
)

type ProductRepository interface {
	Create(ctx context.Context, p repository.Product) (repository.Product, error)
	GetByID(ctx context.Context, id string) (repository.Product, error)
	List(ctx context.Context, limit, offset int, category string) ([]repository.Product, int, error)
	Update(ctx context.Context, id, name, description, category string, priceCents int64) (repository.Product, error)
	SoftDelete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (repository.Product, repository.StockMovement, error)
	GetStockMovements(ctx context.Context, productID string, limit int) ([]repository.StockMovement, error)
}

type ProductUsecase struct {
	repo ProductRepository
}

func NewProductUsecase(repo ProductRepository) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (uc *ProductUsecase) CreateProduct(ctx context.Context, p repository.Product) (repository.Product, error) {
	return uc.repo.Create(ctx, p)
}

func (uc *ProductUsecase) GetProduct(ctx context.Context, id string) (repository.Product, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *ProductUsecase) ListProducts(ctx context.Context, pageSize, page int, category string) ([]repository.Product, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, pageSize, (page-1)*pageSize, category)
}

func (uc *ProductUsecase) UpdateProduct(ctx context.Context, id, name, description, category string, priceCents int64) (repository.Product, error) {
	return uc.repo.Update(ctx, id, name, description, category, priceCents)
}

func (uc *ProductUsecase) DeleteProduct(ctx context.Context, id string) error {
	return uc.repo.SoftDelete(ctx, id)
}

func (uc *ProductUsecase) UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (repository.Product, repository.StockMovement, error) {
	return uc.repo.UpdateStock(ctx, productID, delta, reason, referenceID)
}

// GetStockLevel は GetByID と GetStockMovements を組み合わせるユースケースの orchestration。
func (uc *ProductUsecase) GetStockLevel(ctx context.Context, productID string) (repository.Product, []repository.StockMovement, error) {
	p, err := uc.repo.GetByID(ctx, productID)
	if err != nil {
		return repository.Product{}, nil, err
	}
	movements, err := uc.repo.GetStockMovements(ctx, productID, 10)
	if err != nil {
		return repository.Product{}, nil, err
	}
	return p, movements, nil
}
