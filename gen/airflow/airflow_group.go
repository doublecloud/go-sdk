package airflow

import (
	"context"
	"google.golang.org/grpc"
)

// Airflow provides access to "airflow" service of DoubleCloud
type Airflow struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// New Airflow creates instance of Airflow
func NewAirflow(g func(ctx context.Context) (*grpc.ClientConn, error)) *Airflow {
	return &Airflow{g}
}

// Cluster gets ClusterServce client
func (c *Airflow) Cluster() *ClusterServiceClient {
	return &ClusterServiceClient{getConn: c.getConn}
}

// Operation gets OperationService client
func (c *Airflow) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: c.getConn}
}
