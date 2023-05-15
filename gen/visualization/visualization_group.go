package visualization

import (
	"context"

	"google.golang.org/grpc"
)

// Visualization provides access to "visualization" service of DoubleCloud
type Visualization struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewVisualization creates instance of Visualization
func NewVisualization(g func(ctx context.Context) (*grpc.ClientConn, error)) *Visualization {
	return &Visualization{g}
}

// Workbook gets WorkbookService client
func (v *Visualization) Workbook() *WorkbookServiceClient {
	return &WorkbookServiceClient{getConn: v.getConn}
}
