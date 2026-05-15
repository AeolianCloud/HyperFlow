package ippool

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreatePool(name, gateway string, netmask int, dns1, dns2, description string, nodes, addresses []string) (*Pool, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if gateway == "" {
		return nil, fmt.Errorf("gateway is required")
	}
	if netmask <= 0 || netmask > 32 {
		return nil, fmt.Errorf("netmask must be between 1 and 32")
	}
	if net.ParseIP(gateway) == nil || net.ParseIP(gateway).To4() == nil {
		return nil, fmt.Errorf("invalid gateway address: %s", gateway)
	}
	if dns1 != "" {
		if net.ParseIP(dns1) == nil || net.ParseIP(dns1).To4() == nil {
			return nil, fmt.Errorf("invalid dns1 address: %s", dns1)
		}
	}
	if dns2 != "" {
		if net.ParseIP(dns2) == nil || net.ParseIP(dns2).To4() == nil {
			return nil, fmt.Errorf("invalid dns2 address: %s", dns2)
		}
	}

	expanded, err := ParseAndExpandAddresses(addresses)
	if err != nil {
		return nil, fmt.Errorf("invalid addresses: %w", err)
	}

	for _, addr := range expanded {
		conflict, err := s.store.CheckAddressConflict(addr, "")
		if err != nil {
			return nil, err
		}
		if conflict {
			return nil, fmt.Errorf("address %s already exists in another pool", addr)
		}
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	if nodes == nil {
		nodes = []string{}
	}

	p := &Pool{
		ID:          id,
		Name:        name,
		Gateway:     gateway,
		Netmask:     netmask,
		DNS1:        dns1,
		DNS2:        dns2,
		Description: description,
		Nodes:       nodes,
	}
	if err := s.store.CreatePool(p); err != nil {
		return nil, err
	}

	if len(expanded) > 0 {
		addrs := make([]Address, len(expanded))
		for i, addr := range expanded {
			addrID, _ := generateID()
			addrs[i] = Address{
				ID:      addrID,
				PoolID:  id,
				Address: addr,
			}
		}
		if err := s.store.InsertAddresses(addrs); err != nil {
			return nil, err
		}
	}

	return s.store.GetPool(id)
}

func (s *Service) GetPool(id string) (*Pool, error) {
	return s.store.GetPool(id)
}

func (s *Service) ListPools() ([]*Pool, error) {
	return s.store.ListPools()
}

func (s *Service) UpdatePool(id, name, dns1, dns2, description string, nodes []string) (*Pool, error) {
	existing, err := s.store.GetPool(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("pool not found")
	}

	if dns1 != "" {
		if net.ParseIP(dns1) == nil || net.ParseIP(dns1).To4() == nil {
			return nil, fmt.Errorf("invalid dns1 address: %s", dns1)
		}
	}
	if dns2 != "" {
		if net.ParseIP(dns2) == nil || net.ParseIP(dns2).To4() == nil {
			return nil, fmt.Errorf("invalid dns2 address: %s", dns2)
		}
	}

	if nodes == nil {
		nodes = []string{}
	}

	p := &Pool{
		ID:          id,
		Name:        name,
		DNS1:        dns1,
		DNS2:        dns2,
		Description: description,
		Nodes:       nodes,
	}
	if err := s.store.UpdatePool(p); err != nil {
		return nil, err
	}
	return s.store.GetPool(id)
}

func (s *Service) DeletePool(id string) error {
	hasUsed, err := s.store.HasUsedAddresses(id)
	if err != nil {
		return err
	}
	if hasUsed {
		return fmt.Errorf("cannot delete pool with used or reserved addresses")
	}
	return s.store.DeletePool(id)
}

func (s *Service) InsertAddresses(poolID string, addresses []string) error {
	existing, err := s.store.GetPool(poolID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("pool not found")
	}

	expanded, err := ParseAndExpandAddresses(addresses)
	if err != nil {
		return fmt.Errorf("invalid addresses: %w", err)
	}

	for _, addr := range expanded {
		conflict, err := s.store.CheckAddressConflict(addr, poolID)
		if err != nil {
			return err
		}
		if conflict {
			return fmt.Errorf("address %s already exists in another pool", addr)
		}
	}

	addrs := make([]Address, len(expanded))
	for i, addr := range expanded {
		addrID, _ := generateID()
		addrs[i] = Address{
			ID:      addrID,
			PoolID:  poolID,
			Address: addr,
		}
	}
	return s.store.InsertAddresses(addrs)
}

func (s *Service) DeleteAddresses(poolID string, addresses []string) error {
	return s.store.DeleteAddresses(poolID, addresses)
}

func (s *Service) ListAddresses(poolID, status string, page, size int) ([]Address, int, error) {
	return s.store.ListAddresses(poolID, status, page, size)
}

func (s *Service) AllocateAddress(poolID, address string, vmID int) (*Address, error) {
	return s.store.AllocateAddress(poolID, address, vmID)
}

func (s *Service) AllocateRandomAddress(poolID string, vmID int) (*Address, error) {
	return s.store.AllocateRandomAddress(poolID, vmID)
}

func (s *Service) ReleaseAddressByVMID(vmID int) error {
	return s.store.ReleaseAddressByVMID(vmID)
}

func (s *Service) GetPoolForNode(poolID, node string) (*Pool, error) {
	pool, err := s.store.GetPool(poolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}
	valid := false
	for _, n := range pool.Nodes {
		if n == node {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("pool %s is not bound to node %s", poolID, node)
	}
	return pool, nil
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
