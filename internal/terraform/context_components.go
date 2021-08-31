package terraform

import (
	"fmt"

	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/providers"
	"github.com/hashicorp/terraform/internal/provisioners"
)

// contextComponentFactory is the interface that Context uses
// to initialize various components such as providers and provisioners.
// This factory gets more information than the raw maps using to initialize
// a Context. This information is used for debugging.
type contextComponentFactory interface {
	// ResourceProvider creates a new ResourceProvider with the given type.
	ResourceProvider(typ addrs.Provider) (providers.Interface, error)

	// ResourceProvisioner creates a new ResourceProvisioner with the given
	// type.
	ResourceProvisioner(typ string) (provisioners.Interface, error)
}

// basicComponentFactory just calls a factory from a map directly.
type basicComponentFactory struct {
	providers    map[addrs.Provider]providers.Factory
	provisioners map[string]provisioners.Factory
}

func (c *basicComponentFactory) ResourceProvider(typ addrs.Provider) (providers.Interface, error) {
	f, ok := c.providers[typ]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", typ.String())
	}

	return f()
}

func (c *basicComponentFactory) ResourceProvisioner(typ string) (provisioners.Interface, error) {
	f, ok := c.provisioners[typ]
	if !ok {
		return nil, fmt.Errorf("unknown provisioner %q", typ)
	}

	return f()
}
