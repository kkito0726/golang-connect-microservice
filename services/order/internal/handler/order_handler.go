package handler

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderv1 "github.com/ken/connect-microservice/gen/order/v1"
	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	productv1 "github.com/ken/connect-microservice/gen/product/v1"
	"github.com/ken/connect-microservice/gen/product/v1/productv1connect"
	userv1 "github.com/ken/connect-microservice/gen/user/v1"
	"github.com/ken/connect-microservice/gen/user/v1/userv1connect"
	"github.com/ken/connect-microservice/services/order/internal/repository"
)

type OrderHandler struct {
	repo          *repository.OrderRepository
	productClient productv1connect.ProductServiceClient
	userClient    userv1connect.UserServiceClient
}

var _ orderv1connect.OrderServiceHandler = (*OrderHandler)(nil)

func NewOrderHandler(
	repo *repository.OrderRepository,
	productServiceURL string,
	userServiceURL string,
) *OrderHandler {
	return &OrderHandler{
		repo:          repo,
		productClient: productv1connect.NewProductServiceClient(http.DefaultClient, productServiceURL),
		userClient:    userv1connect.NewUserServiceClient(http.DefaultClient, userServiceURL),
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *connect.Request[orderv1.CreateOrderRequest]) (*connect.Response[orderv1.CreateOrderResponse], error) {
	if req.Msg.UserId == "" || len(req.Msg.Items) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id and items are required"))
	}

	// Verify user exists via user-service
	_, err := h.userClient.GetUser(ctx, connect.NewRequest(&userv1.GetUserRequest{Id: req.Msg.UserId}))
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}

	// Fetch product details and validate stock via product-service
	var orderItems []repository.OrderItem
	var totalCents int64

	for _, item := range req.Msg.Items {
		productResp, err := h.productClient.GetProduct(ctx, connect.NewRequest(&productv1.GetProductRequest{Id: item.ProductId}))
		if err != nil {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("product %s not found: %w", item.ProductId, err))
		}
		product := productResp.Msg.Product

		if product.StockQuantity < item.Quantity {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("insufficient stock for %s: have %d, want %d", product.Name, product.StockQuantity, item.Quantity))
		}

		orderItems = append(orderItems, repository.OrderItem{
			ProductID:      item.ProductId,
			ProductName:    product.Name,
			Quantity:       item.Quantity,
			UnitPriceCents: product.PriceCents,
		})
		totalCents += product.PriceCents * int64(item.Quantity)
	}

	// Deduct stock via product-service
	for _, item := range orderItems {
		_, err := h.productClient.UpdateStock(ctx, connect.NewRequest(&productv1.UpdateStockRequest{
			ProductId: item.ProductID,
			Delta:     -item.Quantity,
			Reason:    productv1.StockChangeReason_STOCK_CHANGE_REASON_SALE,
		}))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("deduct stock for %s: %w", item.ProductID, err))
		}
	}

	order, err := h.repo.Create(ctx, req.Msg.UserId, orderItems, totalCents)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create order: %w", err))
	}

	return connect.NewResponse(&orderv1.CreateOrderResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *connect.Request[orderv1.GetOrderRequest]) (*connect.Response[orderv1.GetOrderResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	order, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&orderv1.GetOrderResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *connect.Request[orderv1.ListOrdersRequest]) (*connect.Response[orderv1.ListOrdersResponse], error) {
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int(req.Msg.Page)
	if page <= 0 {
		page = 1
	}

	status := ""
	if req.Msg.Status != orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		status = orderStatusToString(req.Msg.Status)
	}

	orders, total, err := h.repo.List(ctx, req.Msg.UserId, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoOrders := make([]*orderv1.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = toProtoOrder(o)
	}

	return connect.NewResponse(&orderv1.ListOrdersResponse{
		Orders:     protoOrders,
		TotalCount: int32(total),
	}), nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *connect.Request[orderv1.UpdateOrderStatusRequest]) (*connect.Response[orderv1.UpdateOrderStatusResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	status := orderStatusToString(req.Msg.Status)
	order, err := h.repo.UpdateStatus(ctx, req.Msg.Id, status)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&orderv1.UpdateOrderStatusResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *connect.Request[orderv1.CancelOrderRequest]) (*connect.Response[orderv1.CancelOrderResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	order, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if order.Status != "pending" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("can only cancel pending orders, current status: %s", order.Status))
	}

	// Restore stock via product-service
	for _, item := range order.Items {
		_, err := h.productClient.UpdateStock(ctx, connect.NewRequest(&productv1.UpdateStockRequest{
			ProductId:   item.ProductID,
			Delta:       item.Quantity,
			Reason:      productv1.StockChangeReason_STOCK_CHANGE_REASON_RETURN,
			ReferenceId: order.ID,
		}))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("restore stock for %s: %w", item.ProductID, err))
		}
	}

	order, err = h.repo.UpdateStatus(ctx, req.Msg.Id, "cancelled")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&orderv1.CancelOrderResponse{
		Order: toProtoOrder(order),
	}), nil
}

func toProtoOrder(o repository.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &orderv1.OrderItem{
			Id:             item.ID,
			ProductId:      item.ProductID,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		}
	}

	return &orderv1.Order{
		Id:        o.ID,
		UserId:    o.UserID,
		Status:    stringToOrderStatus(o.Status),
		Items:     items,
		TotalCents: o.TotalCents,
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}

func orderStatusToString(s orderv1.OrderStatus) string {
	switch s {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return "pending"
	case orderv1.OrderStatus_ORDER_STATUS_CONFIRMED:
		return "confirmed"
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return "shipped"
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "pending"
	}
}

func stringToOrderStatus(s string) orderv1.OrderStatus {
	switch s {
	case "pending":
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case "confirmed":
		return orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	case "shipped":
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case "delivered":
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
