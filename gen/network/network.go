// nolint
package network

import (
	"context"

	network "github.com/doublecloud/go-genproto/doublecloud/network/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// NetworkServiceClient is a network.NetworkServiceClient with
// lazy GRPC connection initialization.
type NetworkServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements network.NetworkServiceClient
func (c *NetworkServiceClient) Create(ctx context.Context, in *network.CreateNetworkRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements network.NetworkServiceClient
func (c *NetworkServiceClient) Delete(ctx context.Context, in *network.DeleteNetworkRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements network.NetworkServiceClient
func (c *NetworkServiceClient) Get(ctx context.Context, in *network.GetNetworkRequest, opts ...grpc.CallOption) (*network.Network, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkServiceClient(conn).Get(ctx, in, opts...)
}

// Import implements network.NetworkServiceClient
func (c *NetworkServiceClient) Import(ctx context.Context, in *network.ImportNetworkRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkServiceClient(conn).Import(ctx, in, opts...)
}

// List implements network.NetworkServiceClient
func (c *NetworkServiceClient) List(ctx context.Context, in *network.ListNetworksRequest, opts ...grpc.CallOption) (*network.ListNetworksResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return network.NewNetworkServiceClient(conn).List(ctx, in, opts...)
}

type NetworkIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *NetworkServiceClient
	request *network.ListNetworksRequest

	items []*network.Network
}

func (c *NetworkServiceClient) NetworkIterator(ctx context.Context, req *network.ListNetworksRequest, opts ...grpc.CallOption) *NetworkIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &NetworkIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *NetworkIterator) Next() bool {
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

	it.items = response.Networks
	return len(it.items) > 0
}

func (it *NetworkIterator) Take(size int64) ([]*network.Network, error) {
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

	var result []*network.Network

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *NetworkIterator) TakeAll() ([]*network.Network, error) {
	return it.Take(0)
}

func (it *NetworkIterator) Value() *network.Network {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *NetworkIterator) Error() error {
	return it.err
}
