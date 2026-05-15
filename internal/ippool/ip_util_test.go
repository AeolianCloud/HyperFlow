package ippool

import (
	"testing"
)

func TestParseAndExpandAddresses_Range(t *testing.T) {
	ips, err := ParseAndExpandAddresses([]string{"10.0.0.1-10.0.0.5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 5 {
		t.Fatalf("expected 5 IPs, got %d", len(ips))
	}
	expected := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5"}
	for i, ip := range ips {
		if ip != expected[i] {
			t.Fatalf("expected %s at index %d, got %s", expected[i], i, ip)
		}
	}
}

func TestParseAndExpandAddresses_Single(t *testing.T) {
	ips, err := ParseAndExpandAddresses([]string{"192.168.1.100"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 1 || ips[0] != "192.168.1.100" {
		t.Fatalf("expected [192.168.1.100], got %v", ips)
	}
}

func TestParseAndExpandAddresses_MultipleInputs(t *testing.T) {
	ips, err := ParseAndExpandAddresses([]string{"10.0.0.1-10.0.0.2", "10.0.0.10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 3 {
		t.Fatalf("expected 3 IPs, got %d: %v", len(ips), ips)
	}
	if ips[0] != "10.0.0.1" || ips[1] != "10.0.0.2" || ips[2] != "10.0.0.10" {
		t.Fatalf("unexpected IPs: %v", ips)
	}
}

func TestParseAndExpandAddresses_ExceedsMax(t *testing.T) {
	_, err := ParseAndExpandAddresses([]string{"10.0.0.1-10.0.1.10"})
	if err == nil {
		t.Fatal("expected error for exceeding max addresses")
	}
}

func TestParseAndExpandAddresses_Duplicate(t *testing.T) {
	_, err := ParseAndExpandAddresses([]string{"10.0.0.1", "10.0.0.1"})
	if err == nil {
		t.Fatal("expected error for duplicate IP")
	}
}

func TestParseAndExpandAddresses_InvalidFormat(t *testing.T) {
	_, err := ParseAndExpandAddresses([]string{"not-an-ip"})
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestParseAndExpandAddresses_ReverseRange(t *testing.T) {
	_, err := ParseAndExpandAddresses([]string{"10.0.0.10-10.0.0.5"})
	if err == nil {
		t.Fatal("expected error for reversed range")
	}
}

func TestParseCIDR(t *testing.T) {
	n, err := ParseCIDR("24")
	if err != nil || n != 24 {
		t.Fatalf("expected 24, got %d", n)
	}
}

func TestParseCIDR_Invalid(t *testing.T) {
	_, err := ParseCIDR("33")
	if err == nil {
		t.Fatal("expected error for netmask > 32")
	}
}

func TestParseCIDR_NonNumeric(t *testing.T) {
	_, err := ParseCIDR("abc")
	if err == nil {
		t.Fatal("expected error for non-numeric netmask")
	}
}
