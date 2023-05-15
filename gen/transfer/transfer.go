// nolint
package transfer

import (
	"context"

	transfer "github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// TransferServiceClient is a transfer.TransferServiceClient with
// lazy GRPC connection initialization.
type TransferServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Activate implements transfer.TransferServiceClient
func (c *TransferServiceClient) Activate(ctx context.Context, in *transfer.ActivateTransferRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Activate(ctx, in, opts...)
}

// Create implements transfer.TransferServiceClient
func (c *TransferServiceClient) Create(ctx context.Context, in *transfer.CreateTransferRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Create(ctx, in, opts...)
}

// Deactivate implements transfer.TransferServiceClient
func (c *TransferServiceClient) Deactivate(ctx context.Context, in *transfer.DeactivateTransferRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Deactivate(ctx, in, opts...)
}

// Delete implements transfer.TransferServiceClient
func (c *TransferServiceClient) Delete(ctx context.Context, in *transfer.DeleteTransferRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements transfer.TransferServiceClient
func (c *TransferServiceClient) Get(ctx context.Context, in *transfer.GetTransferRequest, opts ...grpc.CallOption) (*transfer.Transfer, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Get(ctx, in, opts...)
}

// List implements transfer.TransferServiceClient
func (c *TransferServiceClient) List(ctx context.Context, in *transfer.ListTransfersRequest, opts ...grpc.CallOption) (*transfer.ListTransfersResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).List(ctx, in, opts...)
}

type TransferIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *TransferServiceClient
	request *transfer.ListTransfersRequest

	items []*transfer.Transfer
}

func (c *TransferServiceClient) TransferIterator(ctx context.Context, req *transfer.ListTransfersRequest, opts ...grpc.CallOption) *TransferIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &TransferIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *TransferIterator) Next() bool {
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

	it.items = response.Transfers
	return len(it.items) > 0
}

func (it *TransferIterator) Take(size int64) ([]*transfer.Transfer, error) {
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

	var result []*transfer.Transfer

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *TransferIterator) TakeAll() ([]*transfer.Transfer, error) {
	return it.Take(0)
}

func (it *TransferIterator) Value() *transfer.Transfer {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *TransferIterator) Error() error {
	return it.err
}

// Update implements transfer.TransferServiceClient
func (c *TransferServiceClient) Update(ctx context.Context, in *transfer.UpdateTransferRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return transfer.NewTransferServiceClient(conn).Update(ctx, in, opts...)
}
