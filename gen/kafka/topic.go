// nolint
package kafka

import (
	"context"

	kafka "github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
	doublecloud "github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

//revive:disable

// TopicServiceClient is a kafka.TopicServiceClient with
// lazy GRPC connection initialization.
type TopicServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Create implements kafka.TopicServiceClient
func (c *TopicServiceClient) Create(ctx context.Context, in *kafka.CreateTopicRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewTopicServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements kafka.TopicServiceClient
func (c *TopicServiceClient) Delete(ctx context.Context, in *kafka.DeleteTopicRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewTopicServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements kafka.TopicServiceClient
func (c *TopicServiceClient) Get(ctx context.Context, in *kafka.GetTopicRequest, opts ...grpc.CallOption) (*kafka.Topic, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewTopicServiceClient(conn).Get(ctx, in, opts...)
}

// List implements kafka.TopicServiceClient
func (c *TopicServiceClient) List(ctx context.Context, in *kafka.ListTopicsRequest, opts ...grpc.CallOption) (*kafka.ListTopicsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewTopicServiceClient(conn).List(ctx, in, opts...)
}

type TopicIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	client  *TopicServiceClient
	request *kafka.ListTopicsRequest

	items []*kafka.Topic
}

func (c *TopicServiceClient) TopicIterator(ctx context.Context, req *kafka.ListTopicsRequest, opts ...grpc.CallOption) *TopicIterator {
	var pageSize int64
	const defaultPageSize = 1000

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return &TopicIterator{
		ctx:      ctx,
		opts:     opts,
		client:   c,
		request:  req,
		pageSize: pageSize,
	}
}

func (it *TopicIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if len(it.items) > 1 {
		it.items[0] = nil
		it.items = it.items[1:]
		return true
	}
	it.items = nil // consume last item, if any

	if it.started {
		return false
	}
	it.started = true

	response, err := it.client.List(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Topics
	return len(it.items) > 0
}

func (it *TopicIterator) Take(size int64) ([]*kafka.Topic, error) {
	if it.err != nil {
		return nil, it.err
	}

	if size == 0 {
		size = 1 << 32 // something insanely large
	}
	it.requestedSize = size
	defer func() {
		// reset iterator for future calls.
		it.requestedSize = 0
	}()

	var result []*kafka.Topic

	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		result = append(result, it.Value())
	}

	if it.err != nil {
		return nil, it.err
	}

	return result, nil
}

func (it *TopicIterator) TakeAll() ([]*kafka.Topic, error) {
	return it.Take(0)
}

func (it *TopicIterator) Value() *kafka.Topic {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *TopicIterator) Error() error {
	return it.err
}

// Update implements kafka.TopicServiceClient
func (c *TopicServiceClient) Update(ctx context.Context, in *kafka.UpdateTopicRequest, opts ...grpc.CallOption) (*doublecloud.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return kafka.NewTopicServiceClient(conn).Update(ctx, in, opts...)
}
