package transfer

import (
	"context"

	"google.golang.org/grpc"
)

// Transfer provides access to "transfer" service of DoubleCloud
type Transfer struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewTransfer creates instance of Transfer
func NewTransfer(g func(ctx context.Context) (*grpc.ClientConn, error)) *Transfer {
	return &Transfer{g}
}

// Endpoint gets EndpointService client
func (t *Transfer) Endpoint() *EndpointServiceClient {
	return &EndpointServiceClient{getConn: t.getConn}
}

// Operation gets OperationService client
func (t *Transfer) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: t.getConn}
}

// Transfer gets TransferService client
func (t *Transfer) Transfer() *TransferServiceClient {
	return &TransferServiceClient{getConn: t.getConn}
}
