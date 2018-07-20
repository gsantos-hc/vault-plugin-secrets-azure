package azuresecrets

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type azureSecretBackend struct {
	*framework.Backend

	client *client
	lock   sync.RWMutex
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func backend() *azureSecretBackend {
	var b = azureSecretBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: framework.PathAppend(
			pathsRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				pathServicePrincipal(&b),
			},
		),
		Secrets: []*framework.Secret{
			secretServicePrincipal(&b),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}

	return &b
}

// reset clears the backend's Provider
// This is useful when the configuration changes and a new Provider should be
// created with the updated settings.
func (b *azureSecretBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.client = nil
}

func (b *azureSecretBackend) invalidate(ctx context.Context, key string) {
	switch key {
	case "config":
		b.reset()
	}
}

const backendHelp = `
The Azure secrets backend dynamically generates Azure service
principals. The SP credentials have a configurable lease and
are automatically revoked at the end of the lease.

After mounting this backend, credentials to manage Azure resources
must be configured with the "config/" endpoints and policies must be
written using the "roles/" endpoints before any credentials can be
generated.
`
