package client

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	orderv1 "github.com/ken/connect-microservice/gen/order/v1"
	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	"github.com/ken/connect-microservice/services/payment/internal/usecase"
)

// コンパイル時にインターフェース実装を検証する。
var _ usecase.OrderClient = (*ConnectOrderClient)(nil)

// ConnectOrderClient は orderv1connect を usecase.OrderClient に適合させるアダプター。
type ConnectOrderClient struct {
	client orderv1connect.OrderServiceClient
}

func NewConnectOrderClient(baseURL string) *ConnectOrderClient {
	return &ConnectOrderClient{
		client: orderv1connect.NewOrderServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectOrderClient) GetOrder(ctx context.Context, id string) (usecase.OrderInfo, error) {
	resp, err := c.client.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{Id: id}))
	if err != nil {
		return usecase.OrderInfo{}, fmt.Errorf("get order: %w", err)
	}
	o := resp.Msg.Order
	return usecase.OrderInfo{
		ID:         o.Id,
		UserID:     o.UserId,
		TotalCents: o.TotalCents,
	}, nil
}

func (c *ConnectOrderClient) UpdateOrderStatus(ctx context.Context, id, status string) error {
	_, err := c.client.UpdateOrderStatus(ctx, connect.NewRequest(&orderv1.UpdateOrderStatusRequest{
		Id:     id,
		Status: stringToOrderStatus(status),
	}))
	return err
}

func (c *ConnectOrderClient) CancelOrder(ctx context.Context, id string) error {
	_, err := c.client.CancelOrder(ctx, connect.NewRequest(&orderv1.CancelOrderRequest{Id: id}))
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
