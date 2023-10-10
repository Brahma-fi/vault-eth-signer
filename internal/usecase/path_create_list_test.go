package usecase

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestBackend_createKeyManager(t *testing.T) {
	b, _ := newTestBackend(t)

	const (
		testSvc1 = "test1-service"
		testSvc2 = "test2-service"
	)

	// create test1
	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	req.Data["serviceName"] = testSvc2
	storage := req.Storage
	_, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// create test2
	req = logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	req.Storage = storage
	req.Data = map[string]interface{}{
		"serviceName": testSvc1,
		"privateKey":  "3ee65159f7aa057c482b1041f18f37ce90ef5e460cb46fd3fa0c40fbae41c7e1",
	}
	_, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ListOperation, "key-managers")
	req.Storage = storage
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expectedList := &logical.Response{
		Data: map[string]interface{}{
			"keys": []string{testSvc1, testSvc2},
		},
	}

	assert.Equal(t, expectedList, resp)

	// read key-manager by service name
	req = logical.TestRequest(t, logical.ReadOperation, "key-managers/"+testSvc1)
	req.Storage = storage
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expectedKm := &logical.Response{
		Data: map[string]interface{}{
			"service_name": testSvc1,
			"addresses": []string{
				"0xBffc2f3Df75367B0f246aF6Ae42AFf59A33f2704",
			},
		},
	}

	assert.Equal(t, expectedKm, resp)
}

func TestBackend_createKeyManagerFailure1(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	sm := NewStorageMock(0, 1, 0, 0)
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal(t, "failed to put", err.Error())
}

func TestBackend_createKeyManagerFailure2(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	data := map[string]interface{}{
		"privateKey": "abc",
	}
	req.Data = data
	sm := NewStorageMock(0, 1, 0, 0)
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal(t, "privateKey must be a 32-byte hexidecimal string", err.Error())
}

func TestBackend_createKeyManagerFailure3(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	data := map[string]interface{}{
		// use N for the secp256k1 curve to trigger an error
		"privateKey": "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
	}
	req.Data = data
	sm := NewStorageMock(0, 1, 0, 0)
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal(t, "error reconstructing private key from input hex, invalid private key, >=N", err.Error())
}

func TestBackend_listPoliciesFailure1(t *testing.T) {
	b, _ := newTestBackend(t)
	req := logical.TestRequest(t, logical.ListOperation, "key-managers")
	sm := NewStorageMock(0, 0, 0, 0)
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal(t, "failed to list", err.Error())
}
