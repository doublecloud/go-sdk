// nolint
package clickhouse

import (
	"context"

	clickhouse "github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	"google.golang.org/grpc"
)

//revive:disable

// VersionServiceClient is a clickhouse.VersionServiceClient with
// lazy GRPC connection initialization.
type VersionServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// List implements clickhouse.VersionServiceClient
func (c *VersionServiceClient) List(ctx context.Context, in *clickhouse.ListVersionsRequest, opts ...grpc.CallOption) (*clickhouse.ListVersionsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewVersionServiceClient(conn).List(ctx, in, opts...)
}

type VersionIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *VersionServiceClient
	request *clickhouse.ListVersionsRequest

	items []*clickhouse.Version
}

func (c *VersionServiceClient) VersionIterator(ctx context.Context, req *clickhouse.ListVersionsRequest, opts ...grpc.CallOption) *VersionIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &VersionIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *VersionIterator) Next() bool {
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

	it.items = response.Versions
	return len(it.items) > 0
}

func (it *VersionIterator) Take(size int64) ([]*clickhouse.Version, error) {
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

	var result []*clickhouse.Version

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *VersionIterator) TakeAll() ([]*clickhouse.Version, error) {
	return it.Take(0)
}

func (it *VersionIterator) Value() *clickhouse.Version {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *VersionIterator) Error() error {
	return it.err
}
