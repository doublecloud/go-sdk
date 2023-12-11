// nolint
package network

import (
	"context"

	logs "github.com/doublecloud/go-genproto/doublecloud/logs/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// OperationServiceClient is a network.OperationServiceClient with
// lazy GRPC connection initialization.
type OperationServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Get implements network.OperationServiceClient
func (c *OperationServiceClient) Get(ctx context.Context, in *logs.GetOperationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return logs.NewOperationServiceClient(conn).Get(ctx, in, opts...)
}
