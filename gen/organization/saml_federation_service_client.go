package organization

import (
	"context"
	"github.com/doublecloud/go-genproto/doublecloud/organizationmanager/saml/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

type SamlFederationServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

func (c *SamlFederationServiceClient) Create(ctx context.Context, in *saml.CreateFederationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return saml.NewFederationServiceClient(conn).Create(ctx, in, opts...)
}

func (c *SamlFederationServiceClient) Delete(ctx context.Context, in *saml.DeleteFederationRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return saml.NewFederationServiceClient(conn).Delete(ctx, in, opts...)
}

func (c *SamlFederationServiceClient) Get(ctx context.Context, in *saml.GetFederationRequest, opts ...grpc.CallOption) (*saml.Federation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return saml.NewFederationServiceClient(conn).Get(ctx, in, opts...)
}
