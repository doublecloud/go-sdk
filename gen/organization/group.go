package organization

import (
	"context"
	"github.com/doublecloud/go-genproto/doublecloud/organizationmanager/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

type GroupServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

func (c *GroupServiceClient) Create(ctx context.Context, in *organizationmanager.CreateGroupRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupServiceClient(conn).Create(ctx, in, opts...)
}

func (c *GroupServiceClient) Delete(ctx context.Context, in *organizationmanager.DeleteGroupRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupServiceClient(conn).Delete(ctx, in, opts...)
}

func (c *GroupServiceClient) Get(ctx context.Context, in *organizationmanager.GetGroupRequest, opts ...grpc.CallOption) (*organizationmanager.Group, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return organizationmanager.NewGroupServiceClient(conn).Get(ctx, in, opts...)
}
