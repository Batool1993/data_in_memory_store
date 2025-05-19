// server/integration_test.go
package server

import (
	"context"
	"data_storage/server/store_service"
	"net/http/httptest"
	"testing"
	"time"

	"data_storage/client"
	"data_storage/server/adapters"
	"data_storage/server/storage"
)

func TestIntegration_StringAndList(t *testing.T) {
	// 1) In-memory repo
	repo := storage.NewDataRepo(10 * time.Millisecond)

	// 2) Application/service layer with default TTL
	svc := store_service.NewStoreService(repo, 1*time.Second)

	// 3) HTTP handler wiring service
	handler := adapters.NewHandler(svc, "my-secret-token")

	// 4) Spin up test server
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// 5) Client against it
	cli, err := client.NewClient(ts.URL, "my-secret-token")
	if err != nil {
		t.Fatalf("client setup: %v", err)
	}
	ctx := context.Background()

	// --- String flow ---
	if err := cli.SetString(ctx, "foo", "bar", 1*time.Second); err != nil {
		t.Fatalf("SetString failed: %v", err)
	}
	got, err := cli.GetString(ctx, "foo")
	if err != nil {
		t.Fatalf("GetString failed: %v", err)
	}
	if got != "bar" {
		t.Errorf("expected \"bar\", got %q", got)
	}

	// wait for TTL to expire
	time.Sleep(5 * time.Second)
	if _, err := cli.GetString(ctx, "foo"); err == nil {
		t.Error("expected error on expired key, got none")
	}

	// --- List flow ---
	if err := cli.LPush(ctx, "mylist", "a", "b", "c"); err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	val, err := cli.RPop(ctx, "mylist")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if val != "c" {
		t.Errorf("expected \"c\", got %q", val)
	}

	// drain remaining and underflow
	cli.RPop(ctx, "mylist")
	cli.RPop(ctx, "mylist")
	if _, err := cli.RPop(ctx, "mylist"); err == nil {
		t.Error("expected error on empty list, got none")
	}
}
