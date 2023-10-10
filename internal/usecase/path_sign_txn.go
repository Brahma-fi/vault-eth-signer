package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathSignTx(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern:        "key-managers/" + framework.GenericNameRegex("name") + "/txn/sign",
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.signTx,
			},
		},
		HelpSynopsis: "Sign a provided transaction object.",
		HelpDescription: `

    Sign a transaction object with properties conforming to the Ethereum JSON-RPC documentation.

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": {Type: framework.TypeString},
			"address": {
				Type:        framework.TypeString,
				Description: "The address that belongs to a private key in the key-manager.",
			},
			"to": {
				Type:        framework.TypeString,
				Description: "(optional when creating new contract) The contract address the transaction is directed to.",
				Default:     "",
			},
			"data": {
				Type:        framework.TypeString,
				Description: "The compiled code of a contract OR the hash of the invoked method signature and encoded parameters.",
			},
			"input": {
				Type:        framework.TypeString,
				Description: "The compiled code of a contract OR the hash of the invoked method signature and encoded parameters.",
			},
			"value": {
				Type:        framework.TypeString,
				Description: "(optional) Integer of the value sent with this transaction (in wei).",
			},
			"nonce": {
				Type:        framework.TypeString,
				Description: "The transaction nonce.",
			},
			"gas": {
				Type:        framework.TypeString,
				Description: "(optional, default: 90000) Integer of the gas provided for the transaction execution. It will return unused gas",
				Default:     "90000",
			},
			"gasPrice": {
				Type:        framework.TypeString,
				Description: "(optional, default: 0) The gas price for the transaction in wei.",
				Default:     "0",
			},
			"gasFeeCap": {
				Type:        framework.TypeString,
				Description: "(optional) Integer of the gasFeeCap  provided for the transaction execution. It will return unused gas",
			},
			"gasTipCap": {
				Type:        framework.TypeString,
				Description: "(optional) Integer of the gasTipCap provided for the transaction execution. It will return unused gas",
			},
			"chainId": {
				Type:        framework.TypeString,
				Description: "(optional) Chain ID of the target blockchain network. If present, EIP155 signer will be used to sign. If omitted, Homestead signer will be used.",
				Default:     "0",
			},
		},
	}
}

func (b *Backend) signTx(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	from := data.Get("name").(string)

	var txDataToSign []byte
	dataInput := data.Get("data").(string)
	// some client such as go-ethereum uses "input" instead of "data"
	if dataInput == "" {
		dataInput = data.Get("input").(string)
	}
	if len(dataInput) > 2 && dataInput[0:2] != "0x" {
		dataInput = "0x" + dataInput
	}

	txDataToSign, err := hexutil.Decode(dataInput)
	if err != nil {
		b.Logger().Error("Failed to decode payload for the 'data' field", "error", err)
		return nil, err
	}

	keyManager, err := b.retrieveKeyManager(ctx, req, from)
	if err != nil {
		b.Logger().Error("Failed to retrieve the signing keyManager", "address", from, "error", err)
		return nil, fmt.Errorf("error retrieving signing keyManager %s", from)
	}

	if keyManager == nil {
		return nil, fmt.Errorf("signing keyManager %s does not exist", from)
	}

	address := data.Get("address").(string)
	var privateKeyStr string
	for _, keyPairs := range keyManager.KeyPairs {
		if keyPairs.Address == address {
			privateKeyStr = keyPairs.PrivateKey
			break
		}
	}

	if privateKeyStr == "" {
		return nil, errors.New("no private key for the input address")
	}

	amount := validNumber(data.Get("value").(string))
	if amount == nil {
		b.Logger().Error("Invalid amount for the 'value' field", "value", data.Get("value").(string))
		return nil, fmt.Errorf("invalid amount for the 'value' field")
	}

	rawAddressTo := data.Get("to").(string)

	chainId := validNumber(data.Get("chainId").(string))
	if chainId == nil {
		b.Logger().Error("Invalid chainId", "chainId", data.Get("chainId").(string))
		return nil, fmt.Errorf("invalid chainId value")
	}

	gasLimitIn := validNumber(data.Get("gas").(string))
	if gasLimitIn == nil {
		b.Logger().Error("Invalid gas limit", "gas", data.Get("gas").(string))
		return nil, fmt.Errorf("invalid gas limit")
	}
	gasLimit := gasLimitIn.Uint64()

	gasPrice := validNumber(data.Get("gasPrice").(string))
	gasFeeCapStr := data.Get("gasFeeCap").(string)
	gasTipCapStr := data.Get("gasTipCap").(string)

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		b.Logger().Error("Error reconstructing private key from retrieved hex", "error", err)
		return nil, fmt.Errorf("error reconstructing private key from retrieved hex")
	}
	defer zeroKey(privateKey)

	nonceIn := validNumber(data.Get("nonce").(string))
	var nonce uint64
	nonce = nonceIn.Uint64()

	var addressTo *common.Address
	if rawAddressTo != "" {
		addressToTemp := common.HexToAddress(rawAddressTo)
		addressTo = &addressToTemp
	}

	var tx *types.Transaction
	if gasFeeCapStr != "" && gasTipCapStr != "" {
		gasFeeCap := validNumber(data.Get("gasFeeCap").(string))
		gasTipCap := validNumber(data.Get("gasTipCap").(string))
		tx = newTransactionWithDynamicFee(addressTo, nonce, gasFeeCap, gasTipCap, gasLimit, txDataToSign, amount)
	} else {
		tx = newLegacyTransaction(addressTo, nonce, gasPrice, gasLimit, txDataToSign, amount)
	}

	var signer types.Signer
	if big.NewInt(0).Cmp(chainId) == 0 {
		signer = types.HomesteadSigner{}
	} else {
		signer = types.LatestSignerForChainID(chainId)
	}

	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		b.Logger().Error("Failed to sign the transaction object", "error", err)
		return nil, err
	}

	var signedTxBuff bytes.Buffer
	err = signedTx.EncodeRLP(&signedTxBuff)
	if err != nil {
		b.Logger().Error("Failed to encode signedTx RLP", "error", err)
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"txHash":   signedTx.Hash().Hex(),
			"signedTx": hexutil.Encode(signedTxBuff.Bytes()),
		},
	}, nil
}
