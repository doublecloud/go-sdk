package network

import (
	"context"

	"google.golang.org/grpc"
)

// Network provides access to "network" service of DoubleCloud
type Network struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewNetwork creates instance of Network
func NewNetwork(g func(ctx context.Context) (*grpc.ClientConn, error)) *Network {
	return &Network{g}
}

// Operation gets OperationService client
func (n *Network) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: n.getConn}
}

// Network gets NetworkService client
func (n *Network) Network() *NetworkServiceClient {
	return &NetworkServiceClient{getConn: n.getConn}
}

// NetworkConnection gets NetworkConnectionService client
func (n *Network) NetworkConnection() *NetworkConnectionServiceClient {
	return &NetworkConnectionServiceClient{getConn: n.getConn}
}
