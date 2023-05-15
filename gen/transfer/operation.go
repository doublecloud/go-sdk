// nolint
package transfer

import (
	"context"

	transfer "github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// OperationServiceClient is a transfer.OperationServiceClient with
// lazy GRPC connection initialization.
type OperationServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Get implements transfer.OperationServiceClient
func (c *OperationServiceClient) Get(ctx context.Context, in *transfer.GetOperationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewOperationServiceClient(conn).Get(ctx, in, opts...)
}
