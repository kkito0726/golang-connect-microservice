package handler

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderv1 "github.com/ken/connect-microservice/gen/order/v1"
	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	paymentv1 "github.com/ken/connect-microservice/gen/payment/v1"
	"github.com/ken/connect-microservice/gen/payment/v1/paymentv1connect"
	"github.com/ken/connect-microservice/services/payment/internal/repository"
)

type PaymentHandler struct {
	repo        *repository.PaymentRepository
	orderClient orderv1connect.OrderServiceClient
}

var _ paymentv1connect.PaymentServiceHandler = (*PaymentHandler)(nil)

func NewPaymentHandler(repo *repository.PaymentRepository, orderServiceURL string) *PaymentHandler {
	return &PaymentHandler{
		repo:        repo,
		orderClient: orderv1connect.NewOrderServiceClient(http.DefaultClient, orderServiceURL),
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *connect.Request[paymentv1.CreatePaymentRequest]) (*connect.Response[paymentv1.CreatePaymentResponse], error) {
	if req.Msg.OrderId == "" || req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("order_id and user_id are required"))
	}
	if req.Msg.Method == paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("payment method is required"))
	}

	// Fetch order details via order-service
	orderResp, err := h.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{Id: req.Msg.OrderId}))
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("order not found: %w", err))
	}
	order := orderResp.Msg.Order

	if order.UserId != req.Msg.UserId {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("order does not belong to user"))
	}

	// Simulate payment processing (always succeeds for learning purposes)
	payment, err := h.repo.Create(ctx, repository.Payment{
		OrderID:     req.Msg.OrderId,
		UserID:      req.Msg.UserId,
		AmountCents: order.TotalCents,
		Status:      "completed",
		Method:      paymentMethodToString(req.Msg.Method),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create payment: %w", err))
	}

	// Update order status to confirmed via order-service
	_, err = h.orderClient.UpdateOrderStatus(ctx, connect.NewRequest(&orderv1.UpdateOrderStatusRequest{
		Id:     req.Msg.OrderId,
		Status: orderv1.OrderStatus_ORDER_STATUS_CONFIRMED,
	}))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update order status: %w", err))
	}

	return connect.NewResponse(&paymentv1.CreatePaymentResponse{
		Payment: toProtoPayment(payment),
	}), nil
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *connect.Request[paymentv1.GetPaymentRequest]) (*connect.Response[paymentv1.GetPaymentResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	payment, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&paymentv1.GetPaymentResponse{
		Payment: toProtoPayment(payment),
	}), nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *connect.Request[paymentv1.ListPaymentsRequest]) (*connect.Response[paymentv1.ListPaymentsResponse], error) {
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int(req.Msg.Page)
	if page <= 0 {
		page = 1
	}

	payments, total, err := h.repo.List(ctx, req.Msg.OrderId, req.Msg.UserId, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoPayments := make([]*paymentv1.Payment, len(payments))
	for i, p := range payments {
		protoPayments[i] = toProtoPayment(p)
	}

	return connect.NewResponse(&paymentv1.ListPaymentsResponse{
		Payments:   protoPayments,
		TotalCount: int32(total),
	}), nil
}

func (h *PaymentHandler) RefundPayment(ctx context.Context, req *connect.Request[paymentv1.RefundPaymentRequest]) (*connect.Response[paymentv1.RefundPaymentResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	payment, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if payment.Status != "completed" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("can only refund completed payments, current status: %s", payment.Status))
	}

	// Update payment status to refunded
	payment, err = h.repo.UpdateStatus(ctx, req.Msg.Id, "refunded")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Cancel the order via order-service (which restores stock)
	_, err = h.orderClient.CancelOrder(ctx, connect.NewRequest(&orderv1.CancelOrderRequest{Id: payment.OrderID}))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cancel order: %w", err))
	}

	return connect.NewResponse(&paymentv1.RefundPaymentResponse{
		Payment: toProtoPayment(payment),
	}), nil
}

func toProtoPayment(p repository.Payment) *paymentv1.Payment {
	return &paymentv1.Payment{
		Id:          p.ID,
		OrderId:     p.OrderID,
		UserId:      p.UserID,
		AmountCents: p.AmountCents,
		Status:      stringToPaymentStatus(p.Status),
		Method:      stringToPaymentMethod(p.Method),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func paymentMethodToString(m paymentv1.PaymentMethod) string {
	switch m {
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return "credit_card"
	case paymentv1.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER:
		return "bank_transfer"
	case paymentv1.PaymentMethod_PAYMENT_METHOD_WALLET:
		return "wallet"
	default:
		return "unknown"
	}
}

func stringToPaymentStatus(s string) paymentv1.PaymentStatus {
	switch s {
	case "pending":
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case "completed":
		return paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case "failed":
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	case "refunded":
		return paymentv1.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func stringToPaymentMethod(s string) paymentv1.PaymentMethod {
	switch s {
	case "credit_card":
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case "bank_transfer":
		return paymentv1.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER
	case "wallet":
		return paymentv1.PaymentMethod_PAYMENT_METHOD_WALLET
	default:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
