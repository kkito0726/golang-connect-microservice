package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/ken/connect-microservice/gen/product/v1"
	"github.com/ken/connect-microservice/gen/product/v1/productv1connect"
	"github.com/ken/connect-microservice/services/product/internal/repository"
)

type ProductHandler struct {
	repo *repository.ProductRepository
}

var _ productv1connect.ProductServiceHandler = (*ProductHandler)(nil)

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *connect.Request[productv1.CreateProductRequest]) (*connect.Response[productv1.CreateProductResponse], error) {
	if req.Msg.Sku == "" || req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("sku and name are required"))
	}

	p, err := h.repo.Create(ctx, repository.Product{
		SKU:           req.Msg.Sku,
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		PriceCents:    req.Msg.PriceCents,
		StockQuantity: req.Msg.StockQuantity,
		Category:      req.Msg.Category,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, err)
	}

	return connect.NewResponse(&productv1.CreateProductResponse{
		Product: toProtoProduct(p),
	}), nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	p, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&productv1.GetProductResponse{
		Product: toProtoProduct(p),
	}), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *connect.Request[productv1.ListProductsRequest]) (*connect.Response[productv1.ListProductsResponse], error) {
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int(req.Msg.Page)
	if page <= 0 {
		page = 1
	}

	products, total, err := h.repo.List(ctx, pageSize, (page-1)*pageSize, req.Msg.Category)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoProducts := make([]*productv1.Product, len(products))
	for i, p := range products {
		protoProducts[i] = toProtoProduct(p)
	}

	return connect.NewResponse(&productv1.ListProductsResponse{
		Products:   protoProducts,
		TotalCount: int32(total),
	}), nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *connect.Request[productv1.UpdateProductRequest]) (*connect.Response[productv1.UpdateProductResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	p, err := h.repo.Update(ctx, req.Msg.Id, req.Msg.Name, req.Msg.Description, req.Msg.Category, req.Msg.PriceCents)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&productv1.UpdateProductResponse{
		Product: toProtoProduct(p),
	}), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *connect.Request[productv1.DeleteProductRequest]) (*connect.Response[productv1.DeleteProductResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	if err := h.repo.SoftDelete(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&productv1.DeleteProductResponse{}), nil
}

func (h *ProductHandler) UpdateStock(ctx context.Context, req *connect.Request[productv1.UpdateStockRequest]) (*connect.Response[productv1.UpdateStockResponse], error) {
	if req.Msg.ProductId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product_id is required"))
	}

	reason := req.Msg.Reason.String()
	p, mv, err := h.repo.UpdateStock(ctx, req.Msg.ProductId, req.Msg.Delta, reason, req.Msg.ReferenceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}

	return connect.NewResponse(&productv1.UpdateStockResponse{
		Product:  toProtoProduct(p),
		Movement: toProtoStockMovement(mv),
	}), nil
}

func (h *ProductHandler) GetStockLevel(ctx context.Context, req *connect.Request[productv1.GetStockLevelRequest]) (*connect.Response[productv1.GetStockLevelResponse], error) {
	if req.Msg.ProductId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product_id is required"))
	}

	p, err := h.repo.GetByID(ctx, req.Msg.ProductId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	movements, err := h.repo.GetStockMovements(ctx, req.Msg.ProductId, 10)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoMovements := make([]*productv1.StockMovement, len(movements))
	for i, m := range movements {
		protoMovements[i] = toProtoStockMovement(m)
	}

	return connect.NewResponse(&productv1.GetStockLevelResponse{
		ProductId:       p.ID,
		StockQuantity:   p.StockQuantity,
		RecentMovements: protoMovements,
	}), nil
}

func toProtoProduct(p repository.Product) *productv1.Product {
	return &productv1.Product{
		Id:            p.ID,
		Sku:           p.SKU,
		Name:          p.Name,
		Description:   p.Description,
		PriceCents:    p.PriceCents,
		StockQuantity: p.StockQuantity,
		Category:      p.Category,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}

func toProtoStockMovement(m repository.StockMovement) *productv1.StockMovement {
	return &productv1.StockMovement{
		Id:          m.ID,
		ProductId:   m.ProductID,
		Delta:       m.Delta,
		Reason:      productv1.StockChangeReason(productv1.StockChangeReason_value[m.Reason]),
		ReferenceId: m.ReferenceID,
		CreatedAt:   timestamppb.New(m.CreatedAt),
	}
}
