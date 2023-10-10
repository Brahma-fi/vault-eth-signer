package usecase

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestBackend_readKeyManagerFailure1(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.ReadOperation, "key-managers/my-service")
	sm := NewStorageMock(0, 0, 0, 0)
	req.Storage = sm
	resp, err := b.HandleRequest(context.Background(), req)

	assert.Nil(t, resp)
	assert.Equal(t, "failed to get", err.Error())
}

func TestBackend_readKeyManagerFailure2(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.ReadOperation, "key-managers/my-service")
	sm := NewStorageMock(0, 0, 0, 0)
	req.Storage = sm
	resp, err := b.HandleRequest(context.Background(), req)

	assert.Nil(t, resp)
	assert.Equal(t, "failed to get", err.Error())
}

func TestBackend_readKeyManagerFailure3(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.ReadOperation, "key-managers/my-service")
	sm := NewStorageMock(0, 1, 0, 0)
	req.Storage = sm
	resp, _ := b.HandleRequest(context.Background(), req)

	assert.Nil(t, resp)
}

func TestBackend_deleteKeyManagerFailure1(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.DeleteOperation, "key-managers/my-service")
	sm := NewStorageMock(0, 0, 0, 0)
	req.Storage = sm
	resp, err := b.HandleRequest(context.Background(), req)

	assert.Nil(t, resp)
	assert.Equal(t, "failed to get", err.Error())
}

func TestBackend_deleteKeyManagerFailure2(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.DeleteOperation, "key-managers/my-service")
	sm := NewStorageMock(0, 1, 0, 0)
	req.Storage = sm
	resp, err := b.HandleRequest(context.Background(), req)

	assert.Nil(t, resp)
	assert.Nil(t, err)
}
