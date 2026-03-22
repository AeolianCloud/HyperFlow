package pve

import "encoding/json"

// VM 表示 PVE 虚拟机
type VM struct {
	VMID   int     `json:"vmid"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPUs   int     `json:"cpus"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

// VmsService 处理虚拟机相关业务逻辑
type VmsService struct {
	client *Client
}

func NewVmsService(c *Client) *VmsService {
	return &VmsService{client: c}
}

func (s *VmsService) ListVms(node string) ([]VM, error) {
	data, err := s.client.Get("/nodes/" + node + "/qemu")
	if err != nil {
		return nil, err
	}
	var vms []VM
	if err := json.Unmarshal(data, &vms); err != nil {
		return nil, err
	}
	return vms, nil
}

func (s *VmsService) GetVm(node, vmid string) (json.RawMessage, error) {
	return s.client.Get("/nodes/" + node + "/qemu/" + vmid + "/status/current")
}

func (s *VmsService) StartVm(node, vmid string) (json.RawMessage, error) {
	return s.client.Post("/nodes/" + node + "/qemu/" + vmid + "/status/start")
}

func (s *VmsService) StopVm(node, vmid string) (json.RawMessage, error) {
	return s.client.Post("/nodes/" + node + "/qemu/" + vmid + "/status/stop")
}

func (s *VmsService) DeleteVm(node, vmid string) (json.RawMessage, error) {
	return s.client.Delete("/nodes/" + node + "/qemu/" + vmid)
}
