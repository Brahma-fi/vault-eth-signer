package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
)

func newTestBackend(t *testing.T) (logical.Backend, logical.Storage) {
	config := &logical.BackendConfig{
		Logger:      logging.NewVaultLogger(log.Trace),
		System:      &logical.StaticSystemView{},
		StorageView: &logical.InmemStorage{},
		BackendUUID: "test",
	}

	b, err := Factory(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	// Wait for the upgrade to finish
	time.Sleep(time.Second)

	return b, config.StorageView
}

type StorageMock struct {
	switches []int
}

func (s StorageMock) List(_ context.Context, _ string) ([]string, error) {
	if s.switches[0] == 1 {
		return []string{"service1", "service2"}, nil
	} else {
		return nil, errors.New("failed to list")
	}
}

func (s StorageMock) Get(_ context.Context, _ string) (*logical.StorageEntry, error) {
	if s.switches[1] == 2 {
		var entry logical.StorageEntry
		return &entry, nil
	} else if s.switches[1] == 1 {
		return nil, nil
	} else {
		return nil, errors.New("failed to get")
	}
}

func (s StorageMock) Put(_ context.Context, _ *logical.StorageEntry) error {
	return errors.New("failed to put")
}

func (s StorageMock) Delete(_ context.Context, _ string) error {
	return errors.New("failed to delete")
}

func NewStorageMock(a, b, c, d int) StorageMock {
	var sm StorageMock
	sm.switches = []int{a, b, c, d}
	return sm
}
