// nolint
package kafka

import (
	"context"

	kafka "github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// ClusterServiceClient is a kafka.ClusterServiceClient with
// lazy GRPC connection initialization.
type ClusterServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Create(ctx context.Context, in *kafka.CreateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Delete(ctx context.Context, in *kafka.DeleteClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Get(ctx context.Context, in *kafka.GetClusterRequest, opts ...grpc.CallOption) (*kafka.Cluster, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Get(ctx, in, opts...)
}

// List implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) List(ctx context.Context, in *kafka.ListClustersRequest, opts ...grpc.CallOption) (*kafka.ListClustersResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).List(ctx, in, opts...)
}

type ClusterIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *kafka.ListClustersRequest

	items []*kafka.Cluster
}

func (c *ClusterServiceClient) ClusterIterator(ctx context.Context, req *kafka.ListClustersRequest, opts ...grpc.CallOption) *ClusterIterator {
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

func (it *ClusterIterator) Take(size int64) ([]*kafka.Cluster, error) {
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

	var result []*kafka.Cluster

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ClusterIterator) TakeAll() ([]*kafka.Cluster, error) {
	return it.Take(0)
}

func (it *ClusterIterator) Value() *kafka.Cluster {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterIterator) Error() error {
	return it.err
}

// ListHosts implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) ListHosts(ctx context.Context, in *kafka.ListClusterHostsRequest, opts ...grpc.CallOption) (*kafka.ListClusterHostsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).ListHosts(ctx, in, opts...)
}

type ClusterHostsIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *kafka.ListClusterHostsRequest

	items []*kafka.Host
}

func (c *ClusterServiceClient) ClusterHostsIterator(ctx context.Context, req *kafka.ListClusterHostsRequest, opts ...grpc.CallOption) *ClusterHostsIterator {
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

func (it *ClusterHostsIterator) Take(size int64) ([]*kafka.Host, error) {
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

	var result []*kafka.Host

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ClusterHostsIterator) TakeAll() ([]*kafka.Host, error) {
	return it.Take(0)
}

func (it *ClusterHostsIterator) Value() *kafka.Host {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterHostsIterator) Error() error {
	return it.err
}

// ListOperations implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) ListOperations(ctx context.Context, in *kafka.ListClusterOperationsRequest, opts ...grpc.CallOption) (*kafka.ListClusterOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).ListOperations(ctx, in, opts...)
}

type ClusterOperationsIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *kafka.ListClusterOperationsRequest

	items []*doublecloud.Operation
}

func (c *ClusterServiceClient) ClusterOperationsIterator(ctx context.Context, req *kafka.ListClusterOperationsRequest, opts ...grpc.CallOption) *ClusterOperationsIterator {
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

// RescheduleMaintenance implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) RescheduleMaintenance(ctx context.Context, in *kafka.RescheduleMaintenanceRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).RescheduleMaintenance(ctx, in, opts...)
}

// ResetCredentials implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) ResetCredentials(ctx context.Context, in *kafka.ResetClusterCredentialsRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).ResetCredentials(ctx, in, opts...)
}

// Start implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Start(ctx context.Context, in *kafka.StartClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Start(ctx, in, opts...)
}

// Stop implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Stop(ctx context.Context, in *kafka.StopClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Stop(ctx, in, opts...)
}

// Update implements kafka.ClusterServiceClient
func (c *ClusterServiceClient) Update(ctx context.Context, in *kafka.UpdateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewClusterServiceClient(conn).Update(ctx, in, opts...)
}
