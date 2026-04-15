package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderv1 "github.com/ken/connect-microservice/gen/order/v1"
	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	"github.com/ken/connect-microservice/services/order/internal/domain"
	"github.com/ken/connect-microservice/services/order/internal/usecase"
)

type OrderHandler struct {
	uc *usecase.OrderUsecase
}

var _ orderv1connect.OrderServiceHandler = (*OrderHandler)(nil)

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *connect.Request[orderv1.CreateOrderRequest]) (*connect.Response[orderv1.CreateOrderResponse], error) {
	if req.Msg.UserId == "" || len(req.Msg.Items) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id and items are required"))
	}

	items := make([]domain.CreateOrderItem, len(req.Msg.Items))
	for i, item := range req.Msg.Items {
		items[i] = domain.CreateOrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	order, err := h.uc.CreateOrder(ctx, domain.CreateOrderInput{
		UserID: req.Msg.UserId,
		Items:  items,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&orderv1.CreateOrderResponse{Order: toProtoOrder(order)}), nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *connect.Request[orderv1.GetOrderRequest]) (*connect.Response[orderv1.GetOrderResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	order, err := h.uc.GetOrder(ctx, req.Msg.Id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&orderv1.GetOrderResponse{Order: toProtoOrder(order)}), nil
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *connect.Request[orderv1.ListOrdersRequest]) (*connect.Response[orderv1.ListOrdersResponse], error) {
	status := ""
	if req.Msg.Status != orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		status = orderStatusToString(req.Msg.Status)
	}

	orders, total, err := h.uc.ListOrders(ctx, req.Msg.UserId, status, int(req.Msg.PageSize), int(req.Msg.Page))
	if err != nil {
		return nil, toConnectError(err)
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

	order, err := h.uc.UpdateOrderStatus(ctx, req.Msg.Id, orderStatusToString(req.Msg.Status))
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&orderv1.UpdateOrderStatusResponse{Order: toProtoOrder(order)}), nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *connect.Request[orderv1.CancelOrderRequest]) (*connect.Response[orderv1.CancelOrderResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	order, err := h.uc.CancelOrder(ctx, req.Msg.Id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&orderv1.CancelOrderResponse{Order: toProtoOrder(order)}), nil
}

func toProtoOrder(o domain.Order) *orderv1.Order {
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
		Id:         o.ID,
		UserId:     o.UserID,
		Status:     stringToOrderStatus(o.Status),
		Items:      items,
		TotalCents: o.TotalCents,
		CreatedAt:  timestamppb.New(o.CreatedAt),
		UpdatedAt:  timestamppb.New(o.UpdatedAt),
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

func toConnectError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, domain.ErrInsufficientStock):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		slog.Error("internal error", "error", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}
