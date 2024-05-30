package organization

import (
	"context"
	"github.com/doublecloud/go-genproto/doublecloud/organizationmanager/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

type GroupMappingServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

func (c *GroupMappingServiceClient) Create(ctx context.Context, in *organizationmanager.CreateGroupMappingRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupMappingServiceClient(conn).Create(ctx, in, opts...)
}

func (c *GroupMappingServiceClient) Delete(ctx context.Context, in *organizationmanager.DeleteGroupMappingRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupMappingServiceClient(conn).Delete(ctx, in, opts...)
}

func (c *GroupMappingServiceClient) Get(ctx context.Context, in *organizationmanager.GetGroupMappingRequest, opts ...grpc.CallOption) (*organizationmanager.GetGroupMappingResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupMappingServiceClient(conn).Get(ctx, in, opts...)
}
