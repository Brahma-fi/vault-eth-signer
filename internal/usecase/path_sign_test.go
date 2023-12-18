package usecase

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestBackend_sign(t *testing.T) {
	b, _ := newTestBackend(t)

	const (
		testSvc          = "test-service"
		privateKeyString = "3ee65159f7aa057c482b1041f18f37ce90ef5e460cb46fd3fa0c40fbae41c7e1"
	)
	privateKey, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	storage := req.Storage
	req.Data = map[string]interface{}{
		"serviceName": testSvc,
		"privateKey":  privateKeyString,
	}
	_, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// sign contract creation TX by address using Homestead signer
	dataToSign := []byte("data that should be signed")
	hash := crypto.Keccak256Hash(dataToSign)
	req = logical.TestRequest(t, logical.CreateOperation, "key-managers/"+testSvc+"/sign")
	req.Storage = storage
	data := map[string]interface{}{
		"hash":    hash.Hex(),
		"address": address.String(),
	}
	req.Data = data
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	sig := resp.Data["signature"].(string)

	publicKeyBytes := common.Hex2Bytes("045809f2cb46e0a05b7e535e765dc3c658d2a196170f80570900483a46c7875720a2a885656d77181d1107bee5b2f2758a5be3fe58037693c10e7adf16746367bc")
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), common.Hex2Bytes(sig))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, sigPublicKey, publicKeyBytes)
}
