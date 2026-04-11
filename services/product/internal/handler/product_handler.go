package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/ken/connect-microservice/gen/product/v1"
	"github.com/ken/connect-microservice/gen/product/v1/productv1connect"
	"github.com/ken/connect-microservice/services/product/internal/domain"
	"github.com/ken/connect-microservice/services/product/internal/usecase"
)

type ProductHandler struct {
	uc *usecase.ProductUsecase
}

var _ productv1connect.ProductServiceHandler = (*ProductHandler)(nil)

func NewProductHandler(uc *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *connect.Request[productv1.CreateProductRequest]) (*connect.Response[productv1.CreateProductResponse], error) {
	if req.Msg.Sku == "" || req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("sku and name are required"))
	}

	p, err := h.uc.CreateProduct(ctx, domain.Product{
		SKU:           req.Msg.Sku,
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		PriceCents:    req.Msg.PriceCents,
		StockQuantity: req.Msg.StockQuantity,
		Category:      req.Msg.Category,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productv1.CreateProductResponse{Product: toProtoProduct(p)}), nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	p, err := h.uc.GetProduct(ctx, req.Msg.Id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productv1.GetProductResponse{Product: toProtoProduct(p)}), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *connect.Request[productv1.ListProductsRequest]) (*connect.Response[productv1.ListProductsResponse], error) {
	products, total, err := h.uc.ListProducts(ctx, int(req.Msg.PageSize), int(req.Msg.Page), req.Msg.Category)
	if err != nil {
		return nil, toConnectError(err)
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

	p, err := h.uc.UpdateProduct(ctx, req.Msg.Id, req.Msg.Name, req.Msg.Description, req.Msg.Category, req.Msg.PriceCents)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productv1.UpdateProductResponse{Product: toProtoProduct(p)}), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *connect.Request[productv1.DeleteProductRequest]) (*connect.Response[productv1.DeleteProductResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	if err := h.uc.DeleteProduct(ctx, req.Msg.Id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&productv1.DeleteProductResponse{}), nil
}

func (h *ProductHandler) UpdateStock(ctx context.Context, req *connect.Request[productv1.UpdateStockRequest]) (*connect.Response[productv1.UpdateStockResponse], error) {
	if req.Msg.ProductId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product_id is required"))
	}

	p, mv, err := h.uc.UpdateStock(ctx, req.Msg.ProductId, req.Msg.Delta, req.Msg.Reason.String(), req.Msg.ReferenceId)
	if err != nil {
		return nil, toConnectError(err)
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

	p, movements, err := h.uc.GetStockLevel(ctx, req.Msg.ProductId)
	if err != nil {
		return nil, toConnectError(err)
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

func toProtoProduct(p domain.Product) *productv1.Product {
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

func toProtoStockMovement(m domain.StockMovement) *productv1.StockMovement {
	return &productv1.StockMovement{
		Id:          m.ID,
		ProductId:   m.ProductID,
		Delta:       m.Delta,
		Reason:      productv1.StockChangeReason(productv1.StockChangeReason_value[m.Reason]),
		ReferenceId: m.ReferenceID,
		CreatedAt:   timestamppb.New(m.CreatedAt),
	}
}

func toConnectError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, domain.ErrInsufficientStock):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		slog.Error("internal error", "error", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}
