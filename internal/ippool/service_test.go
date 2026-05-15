package ippool

import (
	"testing"
)

func TestService_CreatePool_Success(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, err := svc.CreatePool("test-pool", "192.168.1.1", 24, "8.8.8.8", "", "", []string{"node1"}, []string{"192.168.1.10-192.168.1.12"})
	if err != nil {
		t.Fatalf("CreatePool failed: %v", err)
	}
	if pool.Name != "test-pool" {
		t.Fatalf("expected name test-pool, got %s", pool.Name)
	}
	if pool.Gateway != "192.168.1.1" || pool.Netmask != 24 {
		t.Fatalf("unexpected gateway/netmask: %s/%d", pool.Gateway, pool.Netmask)
	}
}

func TestService_CreatePool_DuplicateName(t *testing.T) {
	svc := NewService(newFakeStore())
	_, err := svc.CreatePool("pool1", "192.168.1.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.1"})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err = svc.CreatePool("pool1", "192.168.2.1", 24, "", "", "", []string{"node2"}, []string{"10.0.0.2"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestService_CreatePool_InvalidInputs(t *testing.T) {
	svc := NewService(newFakeStore())
	tests := []struct {
		name    string
		gateway string
		netmask int
	}{
		{"", "192.168.1.1", 24},
		{"pool1", "", 24},
		{"pool1", "192.168.1.1", 0},
		{"pool1", "192.168.1.1", 33},
		{"pool1", "not-an-ip", 24},
	}
	for _, tt := range tests {
		_, err := svc.CreatePool(tt.name, tt.gateway, tt.netmask, "", "", "", nil, nil)
		if err == nil {
			t.Fatalf("expected error for input: %+v", tt)
		}
	}
}

func TestService_GetPool_NotFound(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, err := svc.GetPool("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool != nil {
		t.Fatal("expected nil pool")
	}
}

func TestService_UpdatePool_OnlyAllowedFields(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, err := svc.CreatePool("pool1", "10.0.0.1", 24, "8.8.8.8", "", "", []string{"node1"}, []string{"10.0.0.10"})
	if err != nil {
		t.Fatalf("CreatePool failed: %v", err)
	}

	updated, err := svc.UpdatePool(pool.ID, "pool1-renamed", "1.1.1.1", "8.8.4.4", "updated desc", []string{"node1", "node2"})
	if err != nil {
		t.Fatalf("UpdatePool failed: %v", err)
	}
	if updated.Name != "pool1-renamed" || updated.DNS1 != "1.1.1.1" || updated.DNS2 != "8.8.4.4" {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}
}

func TestService_DeletePool_WithUsedAddresses(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, err := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10"})
	if err != nil {
		t.Fatalf("CreatePool failed: %v", err)
	}

	svc.AllocateAddress(pool.ID, "10.0.0.10", 100)

	err = svc.DeletePool(pool.ID)
	if err == nil {
		t.Fatal("expected error when deleting pool with used addresses")
	}
}

func TestService_DeletePool_NoUsedAddresses(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, err := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10"})
	if err != nil {
		t.Fatalf("CreatePool failed: %v", err)
	}
	if err := svc.DeletePool(pool.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AllocateAddress_Specific(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, _ := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10-10.0.0.12"})

	addr, err := svc.AllocateAddress(pool.ID, "10.0.0.11", 100)
	if err != nil {
		t.Fatalf("AllocateAddress failed: %v", err)
	}
	if addr == nil || addr.Address != "10.0.0.11" {
		t.Fatalf("unexpected allocation: %v", addr)
	}

	// same address should fail
	dup, _ := svc.AllocateAddress(pool.ID, "10.0.0.11", 101)
	if dup != nil {
		t.Fatal("expected nil for duplicate allocation")
	}
}

func TestService_AllocateRandomAddress(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, _ := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10-10.0.0.12"})

	addr, err := svc.AllocateRandomAddress(pool.ID, 100)
	if err != nil {
		t.Fatalf("AllocateRandomAddress failed: %v", err)
	}
	if addr == nil || addr.Status != "reserved" {
		t.Fatal("expected reserved address")
	}
}

func TestService_AllocateRandomAddress_Exhausted(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, _ := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10"})

	svc.AllocateRandomAddress(pool.ID, 100)

	addr, _ := svc.AllocateRandomAddress(pool.ID, 101)
	if addr != nil {
		t.Fatal("expected nil when pool is exhausted")
	}
}

func TestService_ReleaseAddressByVMID(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, _ := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1"}, []string{"10.0.0.10"})

	svc.AllocateRandomAddress(pool.ID, 100)
	if err := svc.ReleaseAddressByVMID(100); err != nil {
		t.Fatalf("ReleaseAddressByVMID failed: %v", err)
	}

	addr, _ := svc.AllocateRandomAddress(pool.ID, 101)
	if addr == nil {
		t.Fatal("expected address to be available after release")
	}
}

func TestService_GetPoolForNode(t *testing.T) {
	svc := NewService(newFakeStore())
	pool, _ := svc.CreatePool("pool1", "10.0.0.1", 24, "", "", "", []string{"node1", "node2"}, []string{"10.0.0.10"})

	got, err := svc.GetPoolForNode(pool.ID, "node1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected pool")
	}

	_, err = svc.GetPoolForNode(pool.ID, "node3")
	if err == nil {
		t.Fatal("expected error for unbound node")
	}
}
