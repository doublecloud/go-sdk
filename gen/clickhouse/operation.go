// nolint
package clickhouse

import (
	"context"

	clickhouse "github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// OperationServiceClient is a clickhouse.OperationServiceClient with
// lazy GRPC connection initialization.
type OperationServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Get implements clickhouse.OperationServiceClient
func (c *OperationServiceClient) Get(ctx context.Context, in *clickhouse.GetOperationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewOperationServiceClient(conn).Get(ctx, in, opts...)
}

// List implements clickhouse.OperationServiceClient
func (c *OperationServiceClient) List(ctx context.Context, in *clickhouse.ListOperationsRequest, opts ...grpc.CallOption) (*clickhouse.ListOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewOperationServiceClient(conn).List(ctx, in, opts...)
}

type OperationIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *OperationServiceClient
	request *clickhouse.ListOperationsRequest

	items []*doublecloud.Operation
}

func (c *OperationServiceClient) OperationIterator(ctx context.Context, req *clickhouse.ListOperationsRequest, opts ...grpc.CallOption) *OperationIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &OperationIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *OperationIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if len(it.items) > 1 {
		it.items[0] = nil
		it.items = it.items[1:]
		return true
	}
	it.items = nil // consume last item, if any

	if it.started {
		return false
	}
	it.started = true

	response, err := it.client.List(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Operations
	return len(it.items) > 0
}

func (it *OperationIterator) Take(size int64) ([]*doublecloud.Operation, error) {
	if it.err != nil {
		return nil, it.err
	}

	if size == 0 {
		size = 1 << 32 // something insanely large
	}
	it.requestedSize = size
	defer func() {
		// reset iterator for future calls.
		it.requestedSize = 0
	}()

	var result []*doublecloud.Operation

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *OperationIterator) TakeAll() ([]*doublecloud.Operation, error) {
	return it.Take(0)
}

func (it *OperationIterator) Value() *doublecloud.Operation {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *OperationIterator) Error() error {
	return it.err
}
