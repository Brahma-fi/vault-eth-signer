package usecase

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathSign(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern:        "key-managers/" + framework.GenericNameRegex("name") + "/sign",
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.sign,
			},
		},
		HelpSynopsis: "Sign a provided transaction object.",
		HelpDescription: `

    Sign a transaction object with properties conforming to the Ethereum JSON-RPC documentation.

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": {Type: framework.TypeString},
			"hash": {
				Type:        framework.TypeString,
				Description: "Hex string of the hash that should be signed.",
				Default:     "",
			},
		},
	}
}

func (b *Backend) sign(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	serviceNameInput, ok := data.Get("name").(string)
	if !ok {
		return nil, errInvalidType
	}

	hashInput, ok := data.Get("hash").(string)
	if !ok {
		return nil, errInvalidType
	}

	keyManager, err := b.retrieveKeyManager(ctx, req, serviceNameInput)
	if err != nil {
		b.Logger().Error("Failed to retrieve the signing keyManager",
			"service_name", serviceNameInput, "error", err)
		return nil, fmt.Errorf("error retrieving signing keyManager %s", serviceNameInput)
	}

	if keyManager == nil {
		return nil, fmt.Errorf("signing keyManager %s does not exist", serviceNameInput)
	}

	if len(keyManager.KeyPairs) == 0 {
		return nil, fmt.Errorf("signing keyManager %s does not have a key pair", serviceNameInput)
	}

	privateKey, err := crypto.HexToECDSA(keyManager.KeyPairs[0].PrivateKey)
	if err != nil {
		b.Logger().Error("Error reconstructing private key from retrieved hex", "error", err)
		return nil, fmt.Errorf("error reconstructing private key from retrieved hex")
	}
	defer zeroKey(privateKey)

	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(common.HexToHash(hashInput).Bytes(), privateKey)
	if err != nil {
		b.Logger().Error("Error signing input hash", "error", err)
		return nil, fmt.Errorf("error reconstructing private key from retrieved hex")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"signature": common.Bytes2Hex(sig),
		},
	}, nil
}
