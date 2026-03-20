package client

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	productv1 "github.com/ken/connect-microservice/gen/product/v1"
	"github.com/ken/connect-microservice/gen/product/v1/productv1connect"
	userv1 "github.com/ken/connect-microservice/gen/user/v1"
	"github.com/ken/connect-microservice/gen/user/v1/userv1connect"
	"github.com/ken/connect-microservice/services/order/internal/usecase"
)

// コンパイル時にインターフェース実装を検証する。
var _ usecase.UserClient = (*ConnectUserClient)(nil)
var _ usecase.ProductClient = (*ConnectProductClient)(nil)

// ConnectUserClient は userv1connect を usecase.UserClient に適合させるアダプター。
type ConnectUserClient struct {
	client userv1connect.UserServiceClient
}

func NewConnectUserClient(baseURL string) *ConnectUserClient {
	return &ConnectUserClient{
		client: userv1connect.NewUserServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectUserClient) ValidateUser(ctx context.Context, id string) error {
	_, err := c.client.GetUser(ctx, connect.NewRequest(&userv1.GetUserRequest{Id: id}))
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	return nil
}

// ConnectProductClient は productv1connect を usecase.ProductClient に適合させるアダプター。
type ConnectProductClient struct {
	client productv1connect.ProductServiceClient
}

func NewConnectProductClient(baseURL string) *ConnectProductClient {
	return &ConnectProductClient{
		client: productv1connect.NewProductServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectProductClient) GetProduct(ctx context.Context, id string) (usecase.ProductInfo, error) {
	resp, err := c.client.GetProduct(ctx, connect.NewRequest(&productv1.GetProductRequest{Id: id}))
	if err != nil {
		return usecase.ProductInfo{}, fmt.Errorf("get product: %w", err)
	}
	p := resp.Msg.Product
	return usecase.ProductInfo{
		ID:            p.Id,
		Name:          p.Name,
		PriceCents:    p.PriceCents,
		StockQuantity: p.StockQuantity,
	}, nil
}

func (c *ConnectProductClient) DeductStock(ctx context.Context, productID string, quantity int32) error {
	_, err := c.client.UpdateStock(ctx, connect.NewRequest(&productv1.UpdateStockRequest{
		ProductId: productID,
		Delta:     -quantity,
		Reason:    productv1.StockChangeReason_STOCK_CHANGE_REASON_SALE,
	}))
	return err
}

func (c *ConnectProductClient) RestoreStock(ctx context.Context, productID string, quantity int32, referenceID string) error {
	_, err := c.client.UpdateStock(ctx, connect.NewRequest(&productv1.UpdateStockRequest{
		ProductId:   productID,
		Delta:       quantity,
		Reason:      productv1.StockChangeReason_STOCK_CHANGE_REASON_RETURN,
		ReferenceId: referenceID,
	}))
	return err
}
