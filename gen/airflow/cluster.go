package airflow

import (
	"context"
	airflow "github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

// ClusterServiceClient is an airflow.ClusterServiceClient with
// lazy GRPC connection initialization.
var _ airflow.ClusterServiceClient = &ClusterServiceClient{}

type ClusterServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) Create(ctx context.Context, in *airflow.CreateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) Delete(ctx context.Context, in *airflow.DeleteClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) Get(ctx context.Context, in *airflow.GetClusterRequest, opts ...grpc.CallOption) (*airflow.Cluster, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).Get(ctx, in, opts...)
}

// Update implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) Update(ctx context.Context, in *airflow.UpdateClusterRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).Update(ctx, in, opts...)
}

// List implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) List(ctx context.Context, in *airflow.ListClustersRequest, opts ...grpc.CallOption) (*airflow.ListClustersResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).List(ctx, in, opts...)
}

type ClusterIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *ClusterServiceClient
	request *airflow.ListClustersRequest

	items []*airflow.Cluster
}

func (c *ClusterServiceClient) ClusterIterator(ctx context.Context, req *airflow.ListClustersRequest, opts ...grpc.CallOption) *ClusterIterator {
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

func (it *ClusterIterator) Take(size int64) ([]*airflow.Cluster, error) {
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

	var result []*airflow.Cluster

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *ClusterIterator) Value() *airflow.Cluster {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *ClusterIterator) Error() error {
	return it.err
}

func (it *ClusterIterator) TakeAll() ([]*airflow.Cluster, error) {
	return it.Take(0)
}

// ListOperations implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) ListOperations(ctx context.Context, in *airflow.ListClusterOperationsRequest, opts ...grpc.CallOption) (*airflow.ListClusterOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).ListOperations(ctx, in, opts...)
}

// RescheduleMaintenance implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) RescheduleMaintenance(ctx context.Context, in *airflow.RescheduleMaintenanceRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).RescheduleMaintenance(ctx, in, opts...)
}

// ListCustomImages implements airflow.ClusterServiceClient
func (c *ClusterServiceClient) ListCustomImages(ctx context.Context, in *airflow.ListCustomImagesRequest, opts ...grpc.CallOption) (*airflow.ListCustomImagesResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return airflow.NewClusterServiceClient(conn).ListCustomImages(ctx, in, opts...)
}
