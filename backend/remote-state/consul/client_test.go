package consul

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/backend/remote-state"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/state/remote"
)

func TestRemoteClient_impl(t *testing.T) {
	var _ remote.Client = new(RemoteClient)
}

func TestRemoteClient(t *testing.T) {
	acctest.RemoteTestPrecheck(t)

	// Get the backend
	b := backend.TestBackendConfig(t, New(), map[string]interface{}{
		"address": "demo.consul.io:80",
		"path":    fmt.Sprintf("tf-unit/%s", time.Now().String()),
	})

	// Test
	remotestate.TestClient(t, b)
}

func TestConsul_stateLock(t *testing.T) {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		t.Log("consul lock tests require CONSUL_HTTP_ADDR")
		t.Skip()
	}

	path := fmt.Sprintf("tf-unit/%s", time.Now().String())

	// create 2 instances to get 2 remote.Clients
	sA, err := backend.TestBackendConfig(t, New(), map[string]interface{}{
		"address": addr,
		"path":    path,
	}).State()
	if err != nil {
		t.Fatal(err)
	}

	sB, err := backend.TestBackendConfig(t, New(), map[string]interface{}{
		"address": addr,
		"path":    path,
	}).State()
	if err != nil {
		t.Fatal(err)
	}

	remote.TestRemoteLocks(t, sA.(*remote.State).Client, sB.(*remote.State).Client)
}
