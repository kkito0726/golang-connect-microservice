package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentv1 "github.com/ken/connect-microservice/gen/payment/v1"
	"github.com/ken/connect-microservice/gen/payment/v1/paymentv1connect"
	"github.com/ken/connect-microservice/services/payment/internal/repository"
	"github.com/ken/connect-microservice/services/payment/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUsecase
}

var _ paymentv1connect.PaymentServiceHandler = (*PaymentHandler)(nil)

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *connect.Request[paymentv1.CreatePaymentRequest]) (*connect.Response[paymentv1.CreatePaymentResponse], error) {
	if req.Msg.OrderId == "" || req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("order_id and user_id are required"))
	}
	if req.Msg.Method == paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("payment method is required"))
	}

	payment, err := h.uc.CreatePayment(ctx, req.Msg.OrderId, req.Msg.UserId, paymentMethodToString(req.Msg.Method))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&paymentv1.CreatePaymentResponse{Payment: toProtoPayment(payment)}), nil
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *connect.Request[paymentv1.GetPaymentRequest]) (*connect.Response[paymentv1.GetPaymentResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	payment, err := h.uc.GetPayment(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&paymentv1.GetPaymentResponse{Payment: toProtoPayment(payment)}), nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *connect.Request[paymentv1.ListPaymentsRequest]) (*connect.Response[paymentv1.ListPaymentsResponse], error) {
	payments, total, err := h.uc.ListPayments(ctx, req.Msg.OrderId, req.Msg.UserId, int(req.Msg.PageSize), int(req.Msg.Page))
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

	payment, err := h.uc.RefundPayment(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	return connect.NewResponse(&paymentv1.RefundPaymentResponse{Payment: toProtoPayment(payment)}), nil
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
