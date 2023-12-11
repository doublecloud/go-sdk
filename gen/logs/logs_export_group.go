package network

import (
	"context"

	"google.golang.org/grpc"
)

// Export provides access to "network" service of DoubleCloud
type Export struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewExport creates instance of Export
func NewExport(g func(ctx context.Context) (*grpc.ClientConn, error)) *Export {
	return &Export{g}
}

// Operation gets OperationService client
func (n *Export) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: n.getConn}
}

// Export gets ExportService client
func (n *Export) Export() *ExportServiceClient {
	return &ExportServiceClient{getConn: n.getConn}
}
