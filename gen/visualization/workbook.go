// nolint
package visualization

import (
	"context"

	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	visualization "github.com/doublecloud/go-genproto/doublecloud/visualization/v1"
	"google.golang.org/grpc"
)

//revive:disable

// WorkbookServiceClient is a visualization.WorkbookServiceClient with
// lazy GRPC connection initialization.
type WorkbookServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// AdviseDatasetFields implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) AdviseDatasetFields(ctx context.Context, in *visualization.AdviseDatasetFieldsRequest, opts ...grpc.CallOption) (*visualization.AdviseDatasetFieldsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).AdviseDatasetFields(ctx, in, opts...)
}

// Create implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) Create(ctx context.Context, in *visualization.CreateWorkbookRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).Create(ctx, in, opts...)
}

// CreateConnection implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) CreateConnection(ctx context.Context, in *visualization.CreateWorkbookConnectionRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).CreateConnection(ctx, in, opts...)
}

// Delete implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) Delete(ctx context.Context, in *visualization.DeleteWorkbookRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).Delete(ctx, in, opts...)
}

// DeleteConnection implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) DeleteConnection(ctx context.Context, in *visualization.DeleteWorkbookConnectionRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).DeleteConnection(ctx, in, opts...)
}

// Get implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) Get(ctx context.Context, in *visualization.GetWorkbookRequest, opts ...grpc.CallOption) (*visualization.GetWorkbookResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).Get(ctx, in, opts...)
}

// GetConnection implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) GetConnection(ctx context.Context, in *visualization.GetWorkbookConnectionRequest, opts ...grpc.CallOption) (*visualization.GetWorkbookConnectionResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).GetConnection(ctx, in, opts...)
}

// Update implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) Update(ctx context.Context, in *visualization.UpdateWorkbookRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).Update(ctx, in, opts...)
}

// UpdateConnection implements visualization.WorkbookServiceClient
func (c *WorkbookServiceClient) UpdateConnection(ctx context.Context, in *visualization.UpdateWorkbookConnectionRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return visualization.NewWorkbookServiceClient(conn).UpdateConnection(ctx, in, opts...)
}
