package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID            string
	SKU           string
	Name          string
	Description   string
	PriceCents    int64
	StockQuantity int32
	Category      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StockMovement struct {
	ID          string
	ProductID   string
	Delta       int32
	Reason      string
	ReferenceID string
	CreatedAt   time.Time
}

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, p Product) (Product, error) {
	var result Product
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (sku, name, description, price_cents, stock_quantity, category)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at`,
		p.SKU, p.Name, p.Description, p.PriceCents, p.StockQuantity, p.Category,
	).Scan(&result.ID, &result.SKU, &result.Name, &result.Description, &result.PriceCents,
		&result.StockQuantity, &result.Category, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		return Product{}, fmt.Errorf("insert product: %w", err)
	}
	return result, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
		 FROM products WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
		&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Product{}, fmt.Errorf("product not found: %s", id)
		}
		return Product{}, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int, category string) ([]Product, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`
	listQuery := `SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
		FROM products WHERE deleted_at IS NULL`

	if category != "" {
		countQuery += ` AND category = '` + category + `'`
		listQuery += ` AND category = $3`
	}

	if category != "" {
		err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND category = $1`, category).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("count products: %w", err)
		}
		rows, err := r.pool.Query(ctx,
			`SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
			 FROM products WHERE deleted_at IS NULL AND category = $3
			 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset, category)
		if err != nil {
			return nil, 0, fmt.Errorf("list products: %w", err)
		}
		defer rows.Close()
		return scanProducts(rows, total)
	}

	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
		 FROM products WHERE deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	return scanProducts(rows, total)
}

func scanProducts(rows pgx.Rows, total int) ([]Product, int, error) {
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
			&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, id, name, description, category string, priceCents int64) (Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`UPDATE products SET name = $1, description = $2, price_cents = $3, category = $4, updated_at = now()
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at`,
		name, description, priceCents, category, id,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
		&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Product{}, fmt.Errorf("product not found: %s", id)
		}
		return Product{}, fmt.Errorf("update product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE products SET deleted_at = now(), updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %s", id)
	}
	return nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (Product, StockMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, StockMovement{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var p Product
	err = tx.QueryRow(ctx,
		`SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
		 FROM products WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, productID,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
		&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Product{}, StockMovement{}, fmt.Errorf("product not found: %s", productID)
		}
		return Product{}, StockMovement{}, fmt.Errorf("lock product: %w", err)
	}

	newQuantity := p.StockQuantity + delta
	if newQuantity < 0 {
		return Product{}, StockMovement{}, fmt.Errorf("insufficient stock: have %d, need %d", p.StockQuantity, -delta)
	}

	err = tx.QueryRow(ctx,
		`UPDATE products SET stock_quantity = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at`,
		newQuantity, productID,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
		&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Product{}, StockMovement{}, fmt.Errorf("update stock: %w", err)
	}

	var refID *string
	if referenceID != "" {
		refID = &referenceID
	}
	var mv StockMovement
	err = tx.QueryRow(ctx,
		`INSERT INTO stock_movements (product_id, delta, reason, reference_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, product_id, delta, reason, COALESCE(reference_id::text, ''), created_at`,
		productID, delta, reason, refID,
	).Scan(&mv.ID, &mv.ProductID, &mv.Delta, &mv.Reason, &mv.ReferenceID, &mv.CreatedAt)
	if err != nil {
		return Product{}, StockMovement{}, fmt.Errorf("insert stock movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Product{}, StockMovement{}, fmt.Errorf("commit transaction: %w", err)
	}
	return p, mv, nil
}

func (r *ProductRepository) GetStockMovements(ctx context.Context, productID string, limit int) ([]StockMovement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, delta, reason, COALESCE(reference_id::text, ''), created_at
		 FROM stock_movements WHERE product_id = $1
		 ORDER BY created_at DESC LIMIT $2`, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("list stock movements: %w", err)
	}
	defer rows.Close()

	var movements []StockMovement
	for rows.Next() {
		var m StockMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Delta, &m.Reason, &m.ReferenceID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan stock movement: %w", err)
		}
		movements = append(movements, m)
	}
	return movements, nil
}
