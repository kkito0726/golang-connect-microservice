package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ken/connect-microservice/services/product/db/sqlc"
	"github.com/ken/connect-microservice/services/product/internal/domain"
)

type ProductRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool, queries: db.New(pool)}
}

var _ domain.ProductRepository = (*ProductRepository)(nil)

func (r *ProductRepository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	row, err := r.queries.CreateProduct(ctx, db.CreateProductParams{
		Sku:           p.SKU,
		Name:          p.Name,
		Description:   nullableString(p.Description),
		PriceCents:    p.PriceCents,
		StockQuantity: p.StockQuantity,
		Category:      nullableString(p.Category),
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("insert product: %w", err)
	}
	return productFromRow(row.ID, row.Sku, row.Name, row.Description, row.Category, row.PriceCents, row.StockQuantity, row.CreatedAt, row.UpdatedAt), nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	row, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, fmt.Errorf("get product %s: %w", id, domain.ErrNotFound)
		}
		return domain.Product{}, fmt.Errorf("get product: %w", err)
	}
	return productFromRow(row.ID, row.Sku, row.Name, row.Description, row.Category, row.PriceCents, row.StockQuantity, row.CreatedAt, row.UpdatedAt), nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int, category string) ([]domain.Product, int, error) {
	if category != "" {
		cat := &category
		total, err := r.queries.CountProductsByCategory(ctx, cat)
		if err != nil {
			return nil, 0, fmt.Errorf("count products: %w", err)
		}
		rows, err := r.queries.ListProductsByCategory(ctx, db.ListProductsByCategoryParams{
			Limit:    int32(limit),
			Offset:   int32(offset),
			Category: cat,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("list products: %w", err)
		}
		products := make([]domain.Product, len(rows))
		for i, row := range rows {
			products[i] = productFromRow(row.ID, row.Sku, row.Name, row.Description, row.Category, row.PriceCents, row.StockQuantity, row.CreatedAt, row.UpdatedAt)
		}
		return products, int(total), nil
	}

	total, err := r.queries.CountProducts(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}
	rows, err := r.queries.ListProducts(ctx, db.ListProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	products := make([]domain.Product, len(rows))
	for i, row := range rows {
		products[i] = productFromRow(row.ID, row.Sku, row.Name, row.Description, row.Category, row.PriceCents, row.StockQuantity, row.CreatedAt, row.UpdatedAt)
	}
	return products, int(total), nil
}

func (r *ProductRepository) Update(ctx context.Context, id, name, description, category string, priceCents int64) (domain.Product, error) {
	row, err := r.queries.UpdateProduct(ctx, db.UpdateProductParams{
		Name:        name,
		Description: nullableString(description),
		PriceCents:  priceCents,
		Category:    nullableString(category),
		ID:          id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, fmt.Errorf("update product %s: %w", id, domain.ErrNotFound)
		}
		return domain.Product{}, fmt.Errorf("update product: %w", err)
	}
	return productFromRow(row.ID, row.Sku, row.Name, row.Description, row.Category, row.PriceCents, row.StockQuantity, row.CreatedAt, row.UpdatedAt), nil
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id string) error {
	affected, err := r.queries.SoftDeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("delete product %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (domain.Product, domain.StockMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	locked, err := qtx.GetProductForUpdate(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.StockMovement{}, fmt.Errorf("update stock %s: %w", productID, domain.ErrNotFound)
		}
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("lock product: %w", err)
	}

	p := productFromRow(locked.ID, locked.Sku, locked.Name, locked.Description, locked.Category, locked.PriceCents, locked.StockQuantity, locked.CreatedAt, locked.UpdatedAt)
	updated, err := p.ApplyStockDelta(delta)
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("update stock: %w", err)
	}

	stockRow, err := qtx.UpdateProductStock(ctx, db.UpdateProductStockParams{
		StockQuantity: updated.StockQuantity,
		ID:            productID,
	})
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("update stock: %w", err)
	}

	smRow, err := qtx.CreateStockMovement(ctx, db.CreateStockMovementParams{
		ProductID:   productID,
		Delta:       delta,
		Reason:      reason,
		ReferenceID: stringToUUID(referenceID),
	})
	if err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("insert stock movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Product{}, domain.StockMovement{}, fmt.Errorf("commit transaction: %w", err)
	}

	result := productFromRow(stockRow.ID, stockRow.Sku, stockRow.Name, stockRow.Description, stockRow.Category, stockRow.PriceCents, stockRow.StockQuantity, stockRow.CreatedAt, stockRow.UpdatedAt)
	return result, toStockMovement(smRow), nil
}

func (r *ProductRepository) GetStockMovements(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
	rows, err := r.queries.GetStockMovements(ctx, db.GetStockMovementsParams{
		ProductID: productID,
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list stock movements: %w", err)
	}

	movements := make([]domain.StockMovement, len(rows))
	for i, row := range rows {
		movements[i] = toStockMovement(row)
	}
	return movements, nil
}

func productFromRow(id, sku, name string, description, category *string, priceCents int64, stockQty int32, createdAt, updatedAt pgtype.Timestamptz) domain.Product {
	return domain.Product{
		ID:            id,
		SKU:           sku,
		Name:          name,
		Description:   deref(description),
		PriceCents:    priceCents,
		StockQuantity: stockQty,
		Category:      deref(category),
		CreatedAt:     createdAt.Time,
		UpdatedAt:     updatedAt.Time,
	}
}

func toStockMovement(m db.StockMovement) domain.StockMovement {
	return domain.StockMovement{
		ID:          m.ID,
		ProductID:   m.ProductID,
		Delta:       m.Delta,
		Reason:      m.Reason,
		ReferenceID: uuidToString(m.ReferenceID),
		CreatedAt:   m.CreatedAt.Time,
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func stringToUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	cleaned := strings.ReplaceAll(s, "-", "")
	if len(cleaned) != 32 {
		return pgtype.UUID{}
	}
	var b [16]byte
	if _, err := hex.Decode(b[:], []byte(cleaned)); err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: b, Valid: true}
}
