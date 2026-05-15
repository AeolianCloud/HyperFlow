package ippool

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
)

type fakeStore struct {
	mu       sync.Mutex
	pools    map[string]*Pool
	addrs    map[string]*Address
	poolIDSeq int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		pools: make(map[string]*Pool),
		addrs: make(map[string]*Address),
	}
}

func (s *fakeStore) CreatePool(p *Pool) error {
	for _, existing := range s.pools {
		if existing.Name == p.Name {
			return errors.New("duplicate name")
		}
	}
	cloned := *p
	cloned.Nodes = append([]string(nil), p.Nodes...)
	s.pools[p.ID] = &cloned
	return nil
}

func (s *fakeStore) GetPool(id string) (*Pool, error) {
	p, ok := s.pools[id]
	if !ok {
		return nil, nil
	}
	cloned := *p
	cloned.Nodes = append([]string(nil), p.Nodes...)

	s.mu.Lock()
	total := 0
	available := 0
	used := 0
	for _, a := range s.addrs {
		if a.PoolID == id {
			total++
			if a.Status == "available" {
				available++
			}
			if a.Status == "used" {
				used++
			}
		}
	}
	s.mu.Unlock()
	cloned.Total = total
	cloned.Available = available
	cloned.Used = used
	return &cloned, nil
}

func (s *fakeStore) ListPools() ([]*Pool, error) {
	var pools []*Pool
	for _, p := range s.pools {
		cloned := *p
		cloned.Nodes = append([]string(nil), p.Nodes...)
		pools = append(pools, &cloned)
	}
	sort.Slice(pools, func(i, j int) bool { return pools[i].Name < pools[j].Name })
	return pools, nil
}

func (s *fakeStore) UpdatePool(p *Pool) error {
	existing, ok := s.pools[p.ID]
	if !ok {
		return errors.New("not found")
	}
	existing.Name = p.Name
	existing.DNS1 = p.DNS1
	existing.DNS2 = p.DNS2
	existing.Description = p.Description
	existing.Nodes = append([]string(nil), p.Nodes...)
	return nil
}

func (s *fakeStore) DeletePool(id string) error {
	delete(s.pools, id)
	for k, a := range s.addrs {
		if a.PoolID == id {
			delete(s.addrs, k)
		}
	}
	return nil
}

func (s *fakeStore) BindNodes(poolID string, nodes []string) error {
	p, ok := s.pools[poolID]
	if !ok {
		return errors.New("not found")
	}
	p.Nodes = append([]string(nil), nodes...)
	return nil
}

func (s *fakeStore) GetPoolNodes(poolID string) ([]string, error) {
	p, ok := s.pools[poolID]
	if !ok {
		return nil, nil
	}
	return append([]string(nil), p.Nodes...), nil
}

func (s *fakeStore) InsertAddresses(addrs []Address) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range addrs {
		a := addrs[i]
		for _, existing := range s.addrs {
			if existing.Address == a.Address {
				return errors.New("duplicate address")
			}
		}
		a.Status = "available"
		cloned := a
		s.addrs[a.ID] = &cloned
	}
	return nil
}

func (s *fakeStore) DeleteAddresses(poolID string, addresses []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, addr := range addresses {
		for id, a := range s.addrs {
			if a.PoolID == poolID && a.Address == addr {
				if a.Status != "available" {
					return errors.New("address not available")
				}
				delete(s.addrs, id)
				break
			}
		}
	}
	return nil
}

func (s *fakeStore) ListAddresses(poolID, status string, page, size int) ([]Address, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []Address
	for _, a := range s.addrs {
		if a.PoolID == poolID {
			if status == "" || a.Status == status {
				filtered = append(filtered, *a)
			}
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Address < filtered[j].Address })
	total := len(filtered)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	start := (page - 1) * size
	if start >= total {
		return []Address{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (s *fakeStore) AllocateAddress(poolID, address string, vmID int) (*Address, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.addrs {
		if a.PoolID == poolID && a.Address == address {
			if a.Status != "available" {
				return nil, nil
			}
			a.Status = "reserved"
			v := vmID
			a.VMID = &v
			cloned := *a
			return &cloned, nil
		}
	}
	return nil, nil
}

func (s *fakeStore) AllocateRandomAddress(poolID string, vmID int) (*Address, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var available []*Address
	for _, a := range s.addrs {
		if a.PoolID == poolID && a.Status == "available" {
			available = append(available, a)
		}
	}
	if len(available) == 0 {
		return nil, nil
	}
	a := available[rand.Intn(len(available))]
	a.Status = "reserved"
	v := vmID
	a.VMID = &v
	cloned := *a
	return &cloned, nil
}

func (s *fakeStore) ReleaseAddressByVMID(vmID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.addrs {
		if a.VMID != nil && *a.VMID == vmID {
			a.Status = "available"
			a.VMID = nil
		}
	}
	return nil
}

func (s *fakeStore) CountByPoolID(poolID string) (total, available, used int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.addrs {
		if a.PoolID == poolID {
			total++
			switch a.Status {
			case "available":
				available++
			case "used":
				used++
			}
		}
	}
	return
}

func (s *fakeStore) HasUsedAddresses(poolID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.addrs {
		if a.PoolID == poolID && (a.Status == "used" || a.Status == "reserved") {
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeStore) CheckAddressConflict(address string, excludePoolID string) (bool, error) {
	for _, a := range s.addrs {
		if a.Address == address && a.PoolID != excludePoolID {
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeStore) GetPoolByNode(node string) ([]*Pool, error) {
	var result []Pool
	for _, p := range s.pools {
		for _, n := range p.Nodes {
			if n == node {
				result = append(result, *p)
			}
		}
	}
	pools := make([]*Pool, len(result))
	for i := range result {
		pools[i] = &result[i]
	}
	return pools, nil
}
