package client

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	orderv1 "github.com/ken/connect-microservice/gen/order/v1"
	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	"github.com/ken/connect-microservice/internal/auth"
	"github.com/ken/connect-microservice/services/payment/internal/domain"
)

// コンパイル時にインターフェース実装を検証する。
var _ domain.OrderClient = (*ConnectOrderClient)(nil)

// withAuthHeader は ctx に保存されたトークンを発信リクエストの Authorization ヘッダに転送する。
func withAuthHeader[T any](ctx context.Context, req *connect.Request[T]) {
	if token, ok := auth.TokenFromContext(ctx); ok {
		req.Header().Set("Authorization", "Bearer "+token)
	}
}

// ConnectOrderClient は orderv1connect を domain.OrderClient に適合させるアダプター。
type ConnectOrderClient struct {
	client orderv1connect.OrderServiceClient
}

func NewConnectOrderClient(baseURL string) *ConnectOrderClient {
	return &ConnectOrderClient{
		client: orderv1connect.NewOrderServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectOrderClient) GetOrder(ctx context.Context, id string) (domain.OrderInfo, error) {
	req := connect.NewRequest(&orderv1.GetOrderRequest{Id: id})
	withAuthHeader(ctx, req)
	resp, err := c.client.GetOrder(ctx, req)
	if err != nil {
		return domain.OrderInfo{}, fmt.Errorf("get order: %w", err)
	}
	o := resp.Msg.Order
	return domain.OrderInfo{
		ID:         o.Id,
		UserID:     o.UserId,
		TotalCents: o.TotalCents,
	}, nil
}

func (c *ConnectOrderClient) UpdateOrderStatus(ctx context.Context, id, status string) error {
	req := connect.NewRequest(&orderv1.UpdateOrderStatusRequest{
		Id:     id,
		Status: stringToOrderStatus(status),
	})
	withAuthHeader(ctx, req)
	_, err := c.client.UpdateOrderStatus(ctx, req)
	return err
}

func (c *ConnectOrderClient) CancelOrder(ctx context.Context, id string) error {
	req := connect.NewRequest(&orderv1.CancelOrderRequest{Id: id})
	withAuthHeader(ctx, req)
	_, err := c.client.CancelOrder(ctx, req)
	return err
}

func stringToOrderStatus(s string) orderv1.OrderStatus {
	switch s {
	case "confirmed":
		return orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	case "shipped":
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case "delivered":
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	}
}
