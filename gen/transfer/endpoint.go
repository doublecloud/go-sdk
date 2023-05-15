// nolint
package transfer

import (
	"context"

	transfer "github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// EndpointServiceClient is a transfer.EndpointServiceClient with
// lazy GRPC connection initialization.
type EndpointServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements transfer.EndpointServiceClient
func (c *EndpointServiceClient) Create(ctx context.Context, in *transfer.CreateEndpointRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewEndpointServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements transfer.EndpointServiceClient
func (c *EndpointServiceClient) Delete(ctx context.Context, in *transfer.DeleteEndpointRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewEndpointServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements transfer.EndpointServiceClient
func (c *EndpointServiceClient) Get(ctx context.Context, in *transfer.GetEndpointRequest, opts ...grpc.CallOption) (*transfer.Endpoint, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewEndpointServiceClient(conn).Get(ctx, in, opts...)
}

// List implements transfer.EndpointServiceClient
func (c *EndpointServiceClient) List(ctx context.Context, in *transfer.ListEndpointsRequest, opts ...grpc.CallOption) (*transfer.ListEndpointsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewEndpointServiceClient(conn).List(ctx, in, opts...)
}

type EndpointIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *EndpointServiceClient
	request *transfer.ListEndpointsRequest

	items []*transfer.Endpoint
}

func (c *EndpointServiceClient) EndpointIterator(ctx context.Context, req *transfer.ListEndpointsRequest, opts ...grpc.CallOption) *EndpointIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &EndpointIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *EndpointIterator) Next() bool {
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

	it.items = response.Endpoints
	return len(it.items) > 0
}

func (it *EndpointIterator) Take(size int64) ([]*transfer.Endpoint, error) {
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

	var result []*transfer.Endpoint

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *EndpointIterator) TakeAll() ([]*transfer.Endpoint, error) {
	return it.Take(0)
}

func (it *EndpointIterator) Value() *transfer.Endpoint {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *EndpointIterator) Error() error {
	return it.err
}

// Update implements transfer.EndpointServiceClient
func (c *EndpointServiceClient) Update(ctx context.Context, in *transfer.UpdateEndpointRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewEndpointServiceClient(conn).Update(ctx, in, opts...)
}
