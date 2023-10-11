package usecase

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

type KeyManager struct {
	ServiceName string     `json:"service_name"`
	KeyPairs    []*KeyPair `json:"key_pairs"`
}

func paths(b *Backend) []*framework.Path {
	return []*framework.Path{
		pathCreateAndList(b),
		pathReadAndDelete(b),
		pathSign(b),
		pathSignTx(b),
	}
}

func (b *Backend) retrieveKeyManager(
	ctx context.Context,
	req *logical.Request,
	serviceName string,
) (*KeyManager, error) {
	path := fmt.Sprintf("key-managers/%s", serviceName)
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		b.Logger().Error("Failed to retrieve the keyManager by service_name", "path", path, "error", err)
		return nil, err
	}

	if entry == nil {
		// could not find the corresponding key for the serviceName
		return nil, nil
	}

	var policy KeyManager
	err = entry.DecodeJSON(&policy)
	if err != nil {
		b.Logger().Error("Failed to decode keyManager", "path", path, "error", err)
		return nil, err
	}
	return &policy, nil
}
