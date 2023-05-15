package operation

import (
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func TestOperation_Metadata_Nil(t *testing.T) {
	op := New(nil, &Proto{})
	actual := op.Metadata()
	assert.Nil(t, actual)
}

func TestOperation_Running(t *testing.T) {
	op := New(nil, &Proto{Status: doublecloud.Operation_STATUS_PENDING})
	assert.False(t, op.Done())
	assert.False(t, op.Ok())
	assert.False(t, op.Failed())
	assert.Nil(t, op.Error())
}

func TestOperation_Ok(t *testing.T) {
	op := New(nil, &Proto{Status: doublecloud.Operation_STATUS_DONE})
	assert.True(t, op.Done())
	assert.True(t, op.Ok())
	assert.False(t, op.Failed())
	assert.Nil(t, op.Error())
}

func TestOperation_Fail(t *testing.T) {
	st := status.Status{Message: "internal error", Code: int32(code.Code_INTERNAL)}
	op := New(nil, &Proto{Status: doublecloud.Operation_STATUS_INVALID, Error: &st})
	assert.True(t, op.Done())
	assert.False(t, op.Ok())
	assert.True(t, op.Failed())
}
