package kafka

import (
	"context"

	"google.golang.org/grpc"
)

// Kafka provides access to "kafka" service of DoubleCloud
type Kafka struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewKafka creates instance of Kafka
func NewKafka(g func(ctx context.Context) (*grpc.ClientConn, error)) *Kafka {
	return &Kafka{g}
}

// Operation gets OperationService client
func (k *Kafka) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: k.getConn}
}

// Cluster gets ClusterService client
func (k *Kafka) Cluster() *ClusterServiceClient {
	return &ClusterServiceClient{getConn: k.getConn}
}

// Topic gets TopicService client
func (k *Kafka) Topic() *TopicServiceClient {
	return &TopicServiceClient{getConn: k.getConn}
}

// User gets UserService client
func (k *Kafka) User() *UserServiceClient {
	return &UserServiceClient{getConn: k.getConn}
}

// Version gets VersionService client
func (k *Kafka) Version() *VersionServiceClient {
	return &VersionServiceClient{getConn: k.getConn}
}
