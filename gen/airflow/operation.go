package airflow

import (
	"context"
	"github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

// OperationServiceClient is an airflow.OperationServiceClient with
// lazy GRPC connection initialization.
var _ airflow.OperationServiceClient = &OperationServiceClient{}

type OperationServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// List implements airflow.OperationServiceClient
func (c *OperationServiceClient) List(ctx context.Context, in *airflow.ListOperationsRequest, opts ...grpc.CallOption) (*airflow.ListOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewOperationServiceClient(conn).List(ctx, in, opts...)
}

// Get implements airflow.OperationServiceClient
func (c *OperationServiceClient) Get(ctx context.Context, in *airflow.GetOperationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewOperationServiceClient(conn).Get(ctx, in, opts...)
}
