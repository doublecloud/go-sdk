// nolint
package network

import (
	"context"

	network "github.com/doublecloud/go-genproto/doublecloud/network/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// NetworkConnectionServiceClient is a network.NetworkConnectionServiceClient with
// lazy GRPC connection initialization.
type NetworkConnectionServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements network.NetworkConnectionServiceClient
func (c *NetworkConnectionServiceClient) Create(ctx context.Context, in *network.CreateNetworkConnectionRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkConnectionServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements network.NetworkConnectionServiceClient
func (c *NetworkConnectionServiceClient) Delete(ctx context.Context, in *network.DeleteNetworkConnectionRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkConnectionServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements network.NetworkConnectionServiceClient
func (c *NetworkConnectionServiceClient) Get(ctx context.Context, in *network.GetNetworkConnectionRequest, opts ...grpc.CallOption) (*network.NetworkConnection, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkConnectionServiceClient(conn).Get(ctx, in, opts...)
}

// List implements network.NetworkConnectionServiceClient
func (c *NetworkConnectionServiceClient) List(ctx context.Context, in *network.ListNetworkConnectionsRequest, opts ...grpc.CallOption) (*network.ListNetworkConnectionsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkConnectionServiceClient(conn).List(ctx, in, opts...)
}

type NetworkConnectionIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *NetworkConnectionServiceClient
	request *network.ListNetworkConnectionsRequest

	items []*network.NetworkConnection
}

func (c *NetworkConnectionServiceClient) NetworkConnectionIterator(ctx context.Context, req *network.ListNetworkConnectionsRequest, opts ...grpc.CallOption) *NetworkConnectionIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &NetworkConnectionIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *NetworkConnectionIterator) Next() bool {
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

	it.items = response.NetworkConnections
	return len(it.items) > 0
}

func (it *NetworkConnectionIterator) Take(size int64) ([]*network.NetworkConnection, error) {
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

	var result []*network.NetworkConnection

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *NetworkConnectionIterator) TakeAll() ([]*network.NetworkConnection, error) {
	return it.Take(0)
}

func (it *NetworkConnectionIterator) Value() *network.NetworkConnection {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *NetworkConnectionIterator) Error() error {
	return it.err
}
