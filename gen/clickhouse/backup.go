// nolint
package clickhouse

import (
	"context"

	clickhouse "github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// BackupServiceClient is a clickhouse.BackupServiceClient with
// lazy GRPC connection initialization.
type BackupServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements clickhouse.BackupServiceClient
func (c *BackupServiceClient) Create(ctx context.Context, in *clickhouse.CreateBackupRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewBackupServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements clickhouse.BackupServiceClient
func (c *BackupServiceClient) Delete(ctx context.Context, in *clickhouse.DeleteBackupRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewBackupServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements clickhouse.BackupServiceClient
func (c *BackupServiceClient) Get(ctx context.Context, in *clickhouse.GetBackupRequest, opts ...grpc.CallOption) (*clickhouse.Backup, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewBackupServiceClient(conn).Get(ctx, in, opts...)
}

// List implements clickhouse.BackupServiceClient
func (c *BackupServiceClient) List(ctx context.Context, in *clickhouse.ListBackupsRequest, opts ...grpc.CallOption) (*clickhouse.ListBackupsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewBackupServiceClient(conn).List(ctx, in, opts...)
}

type BackupIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *BackupServiceClient
	request *clickhouse.ListBackupsRequest

	items []*clickhouse.Backup
}

func (c *BackupServiceClient) BackupIterator(ctx context.Context, req *clickhouse.ListBackupsRequest, opts ...grpc.CallOption) *BackupIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &BackupIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *BackupIterator) Next() bool {
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

	it.items = response.Backups
	return len(it.items) > 0
}

func (it *BackupIterator) Take(size int64) ([]*clickhouse.Backup, error) {
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

	var result []*clickhouse.Backup

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *BackupIterator) TakeAll() ([]*clickhouse.Backup, error) {
	return it.Take(0)
}

func (it *BackupIterator) Value() *clickhouse.Backup {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *BackupIterator) Error() error {
	return it.err
}
