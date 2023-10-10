package usecase

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathReadAndDelete(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern:      "key-managers/" + framework.GenericNameRegex("name"),
		HelpSynopsis: "Create, get or delete a policy by name",
		HelpDescription: `

    GET - return the key-manager by the name
    DELETE - deletes the key-manager by the name

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": {Type: framework.TypeString},
		},
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.readKeyManager,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.deleteKeyManager,
			},
		},
	}
}

func (b *Backend) readKeyManager(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	serviceName := data.Get("name").(string)
	b.Logger().Info("Retrieving key manager for service name", "service_name", serviceName)
	keyManager, err := b.retrieveKeyManager(ctx, req, serviceName)
	if err != nil {
		return nil, err
	}
	if keyManager == nil {
		return nil, fmt.Errorf("keyManager does not exist")
	}

	addresses := make([]string, len(keyManager.KeyPairs))
	for i := range keyManager.KeyPairs {
		addresses[i] = keyManager.KeyPairs[i].Address
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"service_name": keyManager.ServiceName,
			"addresses":    addresses,
		},
	}, nil
}

func (b *Backend) deleteKeyManager(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	serviceName := data.Get("name").(string)
	policy, err := b.retrieveKeyManager(ctx, req, serviceName)
	if err != nil {
		b.Logger().Error("Failed to retrieve the key-manager by service_name",
			"service_name", serviceName, "error", err)
		return nil, err
	}

	if policy == nil {
		return nil, nil
	}

	if err = req.Storage.Delete(ctx, fmt.Sprintf("key-managers/%s", policy.ServiceName)); err != nil {
		b.Logger().Error("Failed to delete the key-manager from storage",
			"service_name", serviceName, "error", err)
		return nil, err
	}
	return nil, nil
}
