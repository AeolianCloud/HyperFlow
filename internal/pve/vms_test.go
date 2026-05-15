package pve

import (
	"testing"
)

func TestNextSCSIIndex_noScsi(t *testing.T) {
	body := map[string]any{"virtio0": "local:0,import-from=...", "net0": "virtio,bridge=vmbr0"}
	idx := NextSCSIIndex(body)
	if idx != 0 {
		t.Fatalf("expected 0, got %d", idx)
	}
}

func TestNextSCSIIndex_sequential(t *testing.T) {
	body := map[string]any{"scsi0": "local:32", "scsi1": "local:100", "scsi2": "local:50"}
	idx := NextSCSIIndex(body)
	if idx != 3 {
		t.Fatalf("expected 3, got %d", idx)
	}
}

func TestNextSCSIIndex_fillHole(t *testing.T) {
	body := map[string]any{"scsi0": "local:32", "scsi2": "local:100"}
	idx := NextSCSIIndex(body)
	if idx != 1 {
		t.Fatalf("expected 1 (fill hole), got %d", idx)
	}
}

func TestNextSCSIIndex_mixed(t *testing.T) {
	body := map[string]any{"scsi0": "local:32", "scsi1": "local:100"}
	idx := NextSCSIIndex(body)
	if idx != 2 {
		t.Fatalf("expected 2, got %d", idx)
	}
}

func TestParseSCSIDisks_empty(t *testing.T) {
	config := map[string]any{"net0": "virtio,bridge=vmbr0"}
	disks := ParseSCSIDisks(config)
	if len(disks) != 0 {
		t.Fatalf("expected 0 disks, got %d", len(disks))
	}
}

func TestParseSCSIDisks_basic(t *testing.T) {
	config := map[string]any{
		"scsi0": "local-lvm:32,import-from=local:import/focal.qcow2",
		"scsi1": "ceph-pool:100",
		"net0":  "virtio,bridge=vmbr0",
	}
	disks := ParseSCSIDisks(config)
	if len(disks) != 2 {
		t.Fatalf("expected 2 disks, got %d", len(disks))
	}
	if disks[0].DiskId != "scsi0" || disks[0].Size != 32 || disks[0].Storage != "local-lvm" {
		t.Fatalf("unexpected disk0: %+v", disks[0])
	}
	if disks[1].DiskId != "scsi1" || disks[1].Size != 100 || disks[1].Storage != "ceph-pool" {
		t.Fatalf("unexpected disk1: %+v", disks[1])
	}
}

func TestParseSCSIDisks_withFormat(t *testing.T) {
	config := map[string]any{
		"scsi0": "ceph-pool:200,format=qcow2",
	}
	disks := ParseSCSIDisks(config)
	if len(disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(disks))
	}
	if disks[0].Format != "qcow2" {
		t.Fatalf("expected format qcow2, got %s", disks[0].Format)
	}
}

func TestParseSCSIDisks_interfaceField(t *testing.T) {
	config := map[string]any{
		"scsi0": "local:50",
	}
	disks := ParseSCSIDisks(config)
	if len(disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(disks))
	}
	if disks[0].Interface != "scsi" {
		t.Fatalf("expected interface scsi, got %s", disks[0].Interface)
	}
}
