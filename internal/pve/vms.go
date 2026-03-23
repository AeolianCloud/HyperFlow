package pve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	DiskSource    string `json:"diskSource" example:"local:import/noble-server-cloudimg-amd64.qcow2"` // 导入磁盘来源，格式 storage:path（必填）
	DiskInterface string `json:"diskInterface,omitempty" example:"scsi0"`                   // 磁盘接口类型，默认 virtio0（可选）
	DiskFormat    string `json:"diskFormat,omitempty" example:"qcow2"`                      // 源磁盘格式，如 qcow2/raw（可选）
	Storage       string `json:"storage" example:"local-lvm"`                               // 目标存储池（必填）
	Network       string `json:"network,omitempty" example:"virtio,bridge=vmbr0"`              // 网络设备配置，格式同 PVE net0，默认 virtio,bridge=vmbr0（可选）
	// CloudInit 配置（可选，使用云镜像时配置首次启动参数）
	CIUser       string `json:"ciUser,omitempty" example:"ubuntu"`                          // CloudInit 登录用户名（可选）
	CIPassword   string `json:"ciPassword,omitempty" example:"secret"`                     // CloudInit 登录密码（可选）
	SSHKeys      string `json:"sshKeys,omitempty" example:"ssh-rsa AAAA..."`               // CloudInit SSH 公钥，多个公钥用换行分隔（可选）
	IPConfig0    string `json:"ipConfig0,omitempty" example:"ip=192.168.1.100/24,gw=192.168.1.1"` // CloudInit 网络配置，格式同 PVE ipconfig0，如 ip=dhcp（可选）
	Nameserver      string   `json:"nameserver,omitempty" example:"8.8.8.8"`                    // CloudInit DNS 服务器地址（可选）
	SearchDomain    string   `json:"searchDomain,omitempty" example:"example.com"`              // CloudInit DNS 搜索域（可选）
	CIPackages      []string `json:"ciPackages,omitempty" example:"qemu-guest-agent"`        // CloudInit 首次开机需安装的软件包列表（可选；非空时须同时指定 snippetsStorage）
	SnippetsStorage string   `json:"snippetsStorage,omitempty" example:"local"`                  // 存放 cloud-init user-data Snippet 的 PVE 存储名称（ciPackages 非空时必填）
	AptMirror      string   `json:"aptMirror,omitempty" example:"http://mirrors.aliyun.com/ubuntu"` // APT 镜像源地址，写入 cloud-init apt.primary（可选，仅 ciPackages 非空时生效）
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
	net := req.Network
	if net == "" {
		net = "virtio,bridge=vmbr0"
	}
	body["net0"] = net
	// 若包含任意 CloudInit 字段，附加 CloudInit 驱动盘及配置参数
	hasCloudInit := req.CIUser != "" || req.CIPassword != "" || req.SSHKeys != "" ||
		req.IPConfig0 != "" || req.Nameserver != "" || req.SearchDomain != "" ||
		len(req.CIPackages) > 0
	if hasCloudInit {
		body["ide2"] = req.Storage + ":cloudinit"
		if len(req.CIPackages) > 0 {
			// ciPackages 非空：生成 user-data Snippet 并通过 cicustom 引用
			// ciUser/ciPassword/sshKeys 已写入 user-data 文件，不再传给 PVE
			if req.SnippetsStorage == "" {
				return nil, fmt.Errorf("snippetsStorage is required when ciPackages is specified")
			}
			snippetName := fmt.Sprintf("cloudinit-%d-userdata.yaml", req.VMID)
			userData := buildCloudInitUserData(req.CIPackages, req.CIUser, req.CIPassword, req.SSHKeys, req.AptMirror)
			if err := s.UploadSnippet(node, req.SnippetsStorage, snippetName, userData); err != nil {
				return nil, err
			}
			body["cicustom"] = "user=" + req.SnippetsStorage + ":snippets/" + snippetName
		} else {
			// 无 ciPackages：使用 PVE 原生 CloudInit 参数
			if req.CIUser != "" {
				body["ciuser"] = req.CIUser
			}
			if req.CIPassword != "" {
				body["cipassword"] = req.CIPassword
			}
			if req.SSHKeys != "" {
				body["sshkeys"] = url.QueryEscape(req.SSHKeys)
			}
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

// buildCloudInitUserData 生成 cloud-init user-data YAML 内容，默认执行软件包更新和升级
func buildCloudInitUserData(packages []string, ciUser, ciPassword, sshKeys, aptMirror string) string {
	var sb strings.Builder
	sb.WriteString("#cloud-config\n")
	if aptMirror != "" {
		sb.WriteString("apt:\n")
		sb.WriteString("  primary:\n")
		sb.WriteString("    - arches: [default]\n")
		sb.WriteString("      uri: " + aptMirror + "\n")
	}
	sb.WriteString("package_update: true\n")
	sb.WriteString("package_upgrade: true\n")
	if len(packages) > 0 {
		sb.WriteString("packages:\n")
		for _, pkg := range packages {
			sb.WriteString("  - " + pkg + "\n")
		}
	}
	if ciUser != "" || ciPassword != "" || sshKeys != "" {
		name := ciUser
		if name == "" {
			name = "ubuntu"
		}
		sb.WriteString("users:\n")
		sb.WriteString("  - name: " + name + "\n")
		sb.WriteString("    groups: sudo,adm\n")
		sb.WriteString("    shell: /bin/bash\n")
		sb.WriteString("    lock_passwd: false\n")
		sb.WriteString("    sudo: ALL=(ALL) NOPASSWD:ALL\n")
		if ciPassword != "" {
			sb.WriteString("    plain_text_passwd: '" + ciPassword + "'\n")
		}
		if sshKeys != "" {
			sb.WriteString("    ssh_authorized_keys:\n")
			for _, key := range strings.Split(strings.TrimSpace(sshKeys), "\n") {
				if key != "" {
					sb.WriteString("      - " + key + "\n")
				}
			}
		}
	}
	if ciPassword != "" {
		sb.WriteString("ssh_pwauth: true\n")
	}
	return sb.String()
}

// UploadSnippet 通过 WebDAV PUT 将 cloud-init user-data 文件上传至 snippets 目录。
// 需在 .env 中配置 PVE_SNIPPETS_WEBDAV_URL，可选配置 PVE_SNIPPETS_WEBDAV_USER / PVE_SNIPPETS_WEBDAV_PASSWORD。
func (s *VmsService) UploadSnippet(node, storage, filename, content string) error {
	baseURL := os.Getenv("PVE_SNIPPETS_WEBDAV_URL")
	if baseURL == "" {
		return fmt.Errorf("PVE_SNIPPETS_WEBDAV_URL is not configured")
	}
	targetURL := strings.TrimRight(baseURL, "/") + "/" + filename
	req, err := http.NewRequest(http.MethodPut, targetURL, strings.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if user := os.Getenv("PVE_SNIPPETS_WEBDAV_USER"); user != "" {
		req.SetBasicAuth(user, os.Getenv("PVE_SNIPPETS_WEBDAV_PASSWORD"))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WebDAV upload failed: %s %s", resp.Status, string(body))
	}
	return nil
}
