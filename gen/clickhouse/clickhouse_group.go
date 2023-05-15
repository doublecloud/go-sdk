package clickhouse

import (
	"context"

	"google.golang.org/grpc"
)

// ClickHouse provides access to "clickhouse" service of DoubleCloud
type ClickHouse struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewClickHouse creates instance of ClickHouse
func NewClickHouse(g func(ctx context.Context) (*grpc.ClientConn, error)) *ClickHouse {
	return &ClickHouse{g}
}

// Backup gets BackupService client
func (c *ClickHouse) Backup() *BackupServiceClient {
	return &BackupServiceClient{getConn: c.getConn}
}

// Operation gets OperationService client
func (c *ClickHouse) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: c.getConn}
}

// Cluster gets ClusterService client
func (c *ClickHouse) Cluster() *ClusterServiceClient {
	return &ClusterServiceClient{getConn: c.getConn}
}

// Version gets VersionService client
func (c *ClickHouse) Version() *VersionServiceClient {
	return &VersionServiceClient{getConn: c.getConn}
}
