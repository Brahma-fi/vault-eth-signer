package usecase

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestBackend_signTx(t *testing.T) {
	b, _ := newTestBackend(t)

	const (
		keeperSvc  = "keeper-service"
		privateKey = "3ee65159f7aa057c482b1041f18f37ce90ef5e460cb46fd3fa0c40fbae41c7e1"
	)

	req := logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	storage := req.Storage
	req.Data = map[string]interface{}{
		"serviceName": keeperSvc,
		"privateKey":  privateKey,
	}
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Add a random key to the same service
	req = logical.TestRequest(t, logical.UpdateOperation, "key-managers")
	req.Storage = storage
	req.Data = map[string]interface{}{
		"serviceName": keeperSvc,
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "key-managers/"+keeperSvc)
	req.Storage = storage
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	assert.Len(t, resp.Data["addresses"], 2)

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal(err)
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// sign contract creation TX by address using Homestead signer
	dataToSign := "608060405234801561001057600080fd5b506040516020806101d783398101604052516000556101a3806100346000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416632a1afcd981146100665780632c46b2051461008d57806360fe47b1146100a25780636d4ce63c1461008d575b600080fd5b34801561007257600080fd5b5061007b6100ba565b60408051918252519081900360200190f35b34801561009957600080fd5b5061007b6100c0565b3480156100ae57600080fd5b5061007b6004356100c6565b60005481565b60005490565b60006064821061013757604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601960248201527f56616c75652063616e206e6f74206265206f7665722031303000000000000000604482015290519081900360640190fd5b60008290556040805183815290517f9455957c3b77d1d4ed071e2b469dd77e37fc5dfd3b4d44dc8a997cc97c7b3d499181900360200190a15050600054905600a165627a7a72305820a22d4674e519555e6f065ccf98b5bd479e108895cbddc10cba200c775d0008730029000000000000000000000000000000000000000000000000000000000000000a"
	req = logical.TestRequest(t, logical.CreateOperation, "key-managers/"+keeperSvc+"/txn/sign")
	req.Storage = storage
	data := map[string]interface{}{
		"address":  "0xBffc2f3Df75367B0f246aF6Ae42AFf59A33f2704",
		"data":     dataToSign,
		"gas":      2000,
		"nonce":    "0x2",
		"gasPrice": 0,
	}
	req.Data = data
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	tx := &types.Transaction{}
	signedTx := resp.Data["signedTx"].(string)
	signatureBytes, err := hexutil.Decode(signedTx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(signatureBytes), 0))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	v, _, _ := tx.RawSignatureValues()
	assert.Equal(t, true, contains([]*big.Int{big.NewInt(27), big.NewInt(28)}, v))

	sender, _ := types.Sender(types.HomesteadSigner{}, tx)
	assert.Equal(t, address.Hex(), sender.Hex())

	// sign TX by address without "0x" using EIP155 signer
	dataToSign = "60fe47b10000000000000000000000000000000000000000000000000000000000000014"
	req = logical.TestRequest(t, logical.CreateOperation, "key-managers/"+keeperSvc+"/txn/sign")
	req.Storage = storage
	data = map[string]interface{}{
		"data":     dataToSign,
		"address":  "0xBffc2f3Df75367B0f246aF6Ae42AFf59A33f2704",
		"to":       "0xf809410b0d6f047c603deb311979cd413e025a84",
		"gas":      2000,
		"nonce":    "0x3",
		"gasPrice": 0,
		"chainId":  "12345",
	}
	req.Data = data
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	tx = &types.Transaction{}
	signedTx = resp.Data["signedTx"].(string)
	signatureBytes, err = hexutil.Decode(signedTx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(signatureBytes), 0))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	v, _, _ = tx.RawSignatureValues()
	assert.True(t, contains([]*big.Int{big.NewInt(24725), big.NewInt(24726)}, v))

	sender, _ = types.Sender(types.LatestSignerForChainID(big.NewInt(12345)), tx)
	assert.Equal(t, address.Hex(), sender.Hex())

	data = map[string]interface{}{
		"input":     dataToSign,
		"address":   "0xBffc2f3Df75367B0f246aF6Ae42AFf59A33f2704",
		"to":        "0xf809410b0d6f047c603deb311979cd413e025a84",
		"gas":       2500,
		"nonce":     "0x3",
		"gasFeeCap": "1",
		"gasTipCap": "0",
		"chainId":   "1",
	}
	req.Data = data
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	tx = &types.Transaction{}
	signedTx = resp.Data["signedTx"].(string)
	signatureBytes, err = hexutil.Decode(signedTx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(signatureBytes), 0))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	v, _, _ = tx.RawSignatureValues()
	assert.True(t, v.Cmp(big.NewInt(1)) == 0)

	sender, _ = types.Sender(types.LatestSignerForChainID(big.NewInt(1)), tx)
	assert.Equal(t, address.Hex(), sender.Hex())

	// sign TX with invalid nonce
	dataToSign = "60fe47b10000000000000000000000000000000000000000000000000000000000000014"
	req = logical.TestRequest(t, logical.CreateOperation, "key-managers/"+keeperSvc+"/txn/sign")
	req.Storage = storage
	data = map[string]interface{}{
		"data":     dataToSign,
		"address":  "0xBffc2f3Df75367B0f246aF6Ae42AFf59A33f2704",
		"to":       "0xf809410b0d6f047c603deb311979cd413e025a84",
		"gas":      2000,
		"nonce":    "0x",
		"gasPrice": 0,
		"chainId":  "12345",
	}
	req.Data = data
	_, err = b.HandleRequest(context.Background(), req)
	assert.ErrorContains(t, err, "invalid nonce")
}
