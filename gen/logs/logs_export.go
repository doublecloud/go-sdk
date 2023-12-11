// nolint
package network

import (
	"context"

	logs "github.com/doublecloud/go-genproto/doublecloud/logs/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// ExportServiceClient is a network.ExportServiceClient with
// lazy GRPC connection initialization.
type ExportServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements network.ExportServiceClient
func (c *ExportServiceClient) Create(ctx context.Context, in *logs.CreateExportRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return logs.NewLogExportServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements network.ExportServiceClient
func (c *ExportServiceClient) Delete(ctx context.Context, in *logs.DeleteExportRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return logs.NewLogExportServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements network.ExportServiceClient
func (c *ExportServiceClient) Get(ctx context.Context, in *logs.GetExportRequest, opts ...grpc.CallOption) (*logs.LogsExport, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return logs.NewLogExportServiceClient(conn).Get(ctx, in, opts...)
}

// List implements network.ExportServiceClient
func (c *ExportServiceClient) List(ctx context.Context, in *logs.ListExportRequest, opts ...grpc.CallOption) (*logs.ListExportResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return logs.NewLogExportServiceClient(conn).List(ctx, in, opts...)
}

type ExportIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ExportServiceClient
	request *logs.ListExportRequest

	items []*logs.LogsExport
}

func (c *ExportServiceClient) ExportIterator(ctx context.Context, req *logs.ListExportRequest, opts ...grpc.CallOption) *ExportIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &ExportIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *ExportIterator) Next() bool {
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

	it.items = response.Exports
	return len(it.items) > 0
}

func (it *ExportIterator) Take(size int64) ([]*logs.LogsExport, error) {
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

	var result []*logs.LogsExport

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ExportIterator) TakeAll() ([]*logs.LogsExport, error) {
	return it.Take(0)
}

func (it *ExportIterator) Value() *logs.LogsExport {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ExportIterator) Error() error {
	return it.err
}
