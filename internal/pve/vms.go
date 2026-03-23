package pve

import (
	"bytes"
	"encoding/json"
	"net/url"
)

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

// CreateVmRequest 新建虚拟机请求参数
type CreateVmRequest struct {
	VMID          int    `json:"vmid" example:"200"`                                        // 新虚拟机 VMID（必填）
	Name          string `json:"name" example:"my-vm"`                                      // 虚拟机名称（必填）
	Cores         int    `json:"cores" example:"2"`                                         // CPU 核数（必填）
	Memory        int    `json:"memory" example:"2048"`                                     // 内存大小，单位 MB（必填）
	DiskSource    string `json:"diskSource" example:"local:import/noble-server-cloudimg-amd64.img.raw"` // 导入磁盘来源，格式 storage:path（必填）
	DiskInterface string `json:"diskInterface,omitempty" example:"scsi0"`                   // 磁盘接口类型，默认 virtio0（可选）
	DiskFormat    string `json:"diskFormat,omitempty" example:"qcow2"`                      // 源磁盘格式，如 qcow2/raw（可选）
	Storage       string `json:"storage" example:"local-lvm"`                               // 目标存储池（必填）
	// CloudInit 配置（可选，使用云镜像时配置首次启动参数）
	CIUser       string `json:"ciUser,omitempty" example:"ubuntu"`                          // CloudInit 登录用户名（可选）
	CIPassword   string `json:"ciPassword,omitempty" example:"secret"`                     // CloudInit 登录密码（可选）
	SSHKeys      string `json:"sshKeys,omitempty" example:"ssh-rsa AAAA..."`               // CloudInit SSH 公钥，多个公钥用换行分隔（可选）
	IPConfig0    string `json:"ipConfig0,omitempty" example:"ip=192.168.1.100/24,gw=192.168.1.1"` // CloudInit 网络配置，格式同 PVE ipconfig0，如 ip=dhcp（可选）
	Nameserver   string `json:"nameserver,omitempty" example:"8.8.8.8"`                    // CloudInit DNS 服务器地址（可选）
	SearchDomain string `json:"searchDomain,omitempty" example:"example.com"`              // CloudInit DNS 搜索域（可选）
}

func (s *VmsService) CreateVm(node string, req CreateVmRequest) (json.RawMessage, error) {
	iface := req.DiskInterface
	if iface == "" {
		iface = "virtio0"
	}
	diskVal := req.Storage + ":0,import-from=" + req.DiskSource
	if req.DiskFormat != "" {
		diskVal += ",format=" + req.DiskFormat
	}
	body := map[string]any{
		"vmid":    req.VMID,
		"name":    req.Name,
		"cores":   req.Cores,
		"memory":  req.Memory,
		"cpu":     "host",
		"machine": "q35",
		"scsihw":  "virtio-scsi-single",
		iface:    diskVal,
	}
	// 若包含任意 CloudInit 字段，附加 CloudInit 驱动盘及配置参数
	hasCloudInit := req.CIUser != "" || req.CIPassword != "" || req.SSHKeys != "" ||
		req.IPConfig0 != "" || req.Nameserver != "" || req.SearchDomain != ""
	if hasCloudInit {
		body["ide2"] = req.Storage + ":cloudinit"
		if req.CIUser != "" {
			body["ciuser"] = req.CIUser
		}
		if req.CIPassword != "" {
			body["cipassword"] = req.CIPassword
		}
		if req.SSHKeys != "" {
			body["sshkeys"] = url.QueryEscape(req.SSHKeys)
		}
		if req.IPConfig0 != "" {
			body["ipconfig0"] = req.IPConfig0
		}
		if req.Nameserver != "" {
			body["nameserver"] = req.Nameserver
		}
		if req.SearchDomain != "" {
			body["searchdomain"] = req.SearchDomain
		}
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return s.client.PostWithBody("/nodes/"+node+"/qemu", bytes.NewReader(bodyBytes))
}
