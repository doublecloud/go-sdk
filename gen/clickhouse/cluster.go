// nolint
package clickhouse

import (
	"context"

	clickhouse "github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// ClusterServiceClient is a clickhouse.ClusterServiceClient with
// lazy GRPC connection initialization.
type ClusterServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Create(ctx context.Context, in *clickhouse.CreateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Delete(ctx context.Context, in *clickhouse.DeleteClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Get(ctx context.Context, in *clickhouse.GetClusterRequest, opts ...grpc.CallOption) (*clickhouse.Cluster, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Get(ctx, in, opts...)
}

// List implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) List(ctx context.Context, in *clickhouse.ListClustersRequest, opts ...grpc.CallOption) (*clickhouse.ListClustersResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).List(ctx, in, opts...)
}

type ClusterIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *clickhouse.ListClustersRequest

	items []*clickhouse.Cluster
}

func (c *ClusterServiceClient) ClusterIterator(ctx context.Context, req *clickhouse.ListClustersRequest, opts ...grpc.CallOption) *ClusterIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &ClusterIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *ClusterIterator) Next() bool {
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

	it.items = response.Clusters
	return len(it.items) > 0
}

func (it *ClusterIterator) Take(size int64) ([]*clickhouse.Cluster, error) {
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

	var result []*clickhouse.Cluster

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ClusterIterator) TakeAll() ([]*clickhouse.Cluster, error) {
	return it.Take(0)
}

func (it *ClusterIterator) Value() *clickhouse.Cluster {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterIterator) Error() error {
	return it.err
}

// ListBackups implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) ListBackups(ctx context.Context, in *clickhouse.ListClusterBackupsRequest, opts ...grpc.CallOption) (*clickhouse.ListClusterBackupsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).ListBackups(ctx, in, opts...)
}

type ClusterBackupsIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *clickhouse.ListClusterBackupsRequest

	items []*clickhouse.Backup
}

func (c *ClusterServiceClient) ClusterBackupsIterator(ctx context.Context, req *clickhouse.ListClusterBackupsRequest, opts ...grpc.CallOption) *ClusterBackupsIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &ClusterBackupsIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *ClusterBackupsIterator) Next() bool {
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

	response, err := it.client.ListBackups(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Backups
	return len(it.items) > 0
}

func (it *ClusterBackupsIterator) Take(size int64) ([]*clickhouse.Backup, error) {
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

func (it *ClusterBackupsIterator) TakeAll() ([]*clickhouse.Backup, error) {
	return it.Take(0)
}

func (it *ClusterBackupsIterator) Value() *clickhouse.Backup {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterBackupsIterator) Error() error {
	return it.err
}

// ListHosts implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) ListHosts(ctx context.Context, in *clickhouse.ListClusterHostsRequest, opts ...grpc.CallOption) (*clickhouse.ListClusterHostsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).ListHosts(ctx, in, opts...)
}

type ClusterHostsIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *clickhouse.ListClusterHostsRequest

	items []*clickhouse.Host
}

func (c *ClusterServiceClient) ClusterHostsIterator(ctx context.Context, req *clickhouse.ListClusterHostsRequest, opts ...grpc.CallOption) *ClusterHostsIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &ClusterHostsIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *ClusterHostsIterator) Next() bool {
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

	response, err := it.client.ListHosts(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Hosts
	return len(it.items) > 0
}

func (it *ClusterHostsIterator) Take(size int64) ([]*clickhouse.Host, error) {
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

	var result []*clickhouse.Host

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ClusterHostsIterator) TakeAll() ([]*clickhouse.Host, error) {
	return it.Take(0)
}

func (it *ClusterHostsIterator) Value() *clickhouse.Host {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterHostsIterator) Error() error {
	return it.err
}

// ListOperations implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) ListOperations(ctx context.Context, in *clickhouse.ListClusterOperationsRequest, opts ...grpc.CallOption) (*clickhouse.ListClusterOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).ListOperations(ctx, in, opts...)
}

type ClusterOperationsIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *clickhouse.ListClusterOperationsRequest

	items []*doublecloud.Operation
}

func (c *ClusterServiceClient) ClusterOperationsIterator(ctx context.Context, req *clickhouse.ListClusterOperationsRequest, opts ...grpc.CallOption) *ClusterOperationsIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &ClusterOperationsIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *ClusterOperationsIterator) Next() bool {
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

	response, err := it.client.ListOperations(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Operations
	return len(it.items) > 0
}

func (it *ClusterOperationsIterator) Take(size int64) ([]*doublecloud.Operation, error) {
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

func (it *ClusterOperationsIterator) TakeAll() ([]*doublecloud.Operation, error) {
	return it.Take(0)
}

func (it *ClusterOperationsIterator) Value() *doublecloud.Operation {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterOperationsIterator) Error() error {
	return it.err
}

// RescheduleMaintenance implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) RescheduleMaintenance(ctx context.Context, in *clickhouse.RescheduleMaintenanceRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).RescheduleMaintenance(ctx, in, opts...)
}

// ResetCredentials implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) ResetCredentials(ctx context.Context, in *clickhouse.ResetClusterCredentialsRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).ResetCredentials(ctx, in, opts...)
}

// Restore implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Restore(ctx context.Context, in *clickhouse.RestoreClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Restore(ctx, in, opts...)
}

// Start implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Start(ctx context.Context, in *clickhouse.StartClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Start(ctx, in, opts...)
}

// Stop implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Stop(ctx context.Context, in *clickhouse.StopClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Stop(ctx, in, opts...)
}

// Update implements clickhouse.ClusterServiceClient
func (c *ClusterServiceClient) Update(ctx context.Context, in *clickhouse.UpdateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return clickhouse.NewClusterServiceClient(conn).Update(ctx, in, opts...)
}
