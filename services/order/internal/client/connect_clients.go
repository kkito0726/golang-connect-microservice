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
	"github.com/ken/connect-microservice/internal/auth"
	"github.com/ken/connect-microservice/services/order/internal/domain"
)

// コンパイル時にインターフェース実装を検証する。
var _ domain.UserClient = (*ConnectUserClient)(nil)
var _ domain.ProductClient = (*ConnectProductClient)(nil)

// withAuthHeader は ctx に保存されたトークンを発信リクエストの Authorization ヘッダに転送する。
func withAuthHeader[T any](ctx context.Context, req *connect.Request[T]) {
	if token, ok := auth.TokenFromContext(ctx); ok {
		req.Header().Set("Authorization", "Bearer "+token)
	}
}

// ConnectUserClient は userv1connect を domain.UserClient に適合させるアダプター。
type ConnectUserClient struct {
	client userv1connect.UserServiceClient
}

func NewConnectUserClient(baseURL string) *ConnectUserClient {
	return &ConnectUserClient{
		client: userv1connect.NewUserServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectUserClient) ValidateUser(ctx context.Context, id string) error {
	req := connect.NewRequest(&userv1.GetUserRequest{Id: id})
	withAuthHeader(ctx, req)
	_, err := c.client.GetUser(ctx, req)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	return nil
}

// ConnectProductClient は productv1connect を domain.ProductClient に適合させるアダプター。
type ConnectProductClient struct {
	client productv1connect.ProductServiceClient
}

func NewConnectProductClient(baseURL string) *ConnectProductClient {
	return &ConnectProductClient{
		client: productv1connect.NewProductServiceClient(http.DefaultClient, baseURL),
	}
}

func (c *ConnectProductClient) GetProduct(ctx context.Context, id string) (domain.ProductInfo, error) {
	req := connect.NewRequest(&productv1.GetProductRequest{Id: id})
	withAuthHeader(ctx, req)
	resp, err := c.client.GetProduct(ctx, req)
	if err != nil {
		return domain.ProductInfo{}, fmt.Errorf("get product: %w", err)
	}
	p := resp.Msg.Product
	return domain.ProductInfo{
		ID:            p.Id,
		Name:          p.Name,
		PriceCents:    p.PriceCents,
		StockQuantity: p.StockQuantity,
	}, nil
}

func (c *ConnectProductClient) DeductStock(ctx context.Context, productID string, quantity int32) error {
	req := connect.NewRequest(&productv1.UpdateStockRequest{
		ProductId: productID,
		Delta:     -quantity,
		Reason:    productv1.StockChangeReason_STOCK_CHANGE_REASON_SALE,
	})
	withAuthHeader(ctx, req)
	_, err := c.client.UpdateStock(ctx, req)
	return err
}

func (c *ConnectProductClient) RestoreStock(ctx context.Context, productID string, quantity int32, referenceID string) error {
	req := connect.NewRequest(&productv1.UpdateStockRequest{
		ProductId:   productID,
		Delta:       quantity,
		Reason:      productv1.StockChangeReason_STOCK_CHANGE_REASON_RETURN,
		ReferenceId: referenceID,
	})
	withAuthHeader(ctx, req)
	_, err := c.client.UpdateStock(ctx, req)
	return err
}
