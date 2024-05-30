package organization

import (
	"context"
	"google.golang.org/grpc"
)

type Organization struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// NewOrganization creates instance of Organization service clients group
func NewOrganization(g func(ctx context.Context) (*grpc.ClientConn, error)) *Organization {
	return &Organization{g}
}

func (n *Organization) Group() *GroupServiceClient {
	return &GroupServiceClient{getConn: n.getConn}
}

func (n *Organization) GroupMapping() *GroupMappingServiceClient {
	return &GroupMappingServiceClient{getConn: n.getConn}
}

func (n *Organization) SamlFederation() *SamlFederationServiceClient {
	return &SamlFederationServiceClient{getConn: n.getConn}
}
