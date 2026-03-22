package pve

import "encoding/json"

// Node 表示 PVE 集群节点
type Node struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

// NodeStatus 表示单个节点的详细状态
type NodeStatus struct {
	CPU     float64 `json:"cpu"`
	Mem     int64   `json:"memory"`
	Uptime  int64   `json:"uptime"`
	KVersion string `json:"kversion"`
}

// NodesService 处理节点相关业务逻辑
type NodesService struct {
	client *Client
}

func NewNodesService(c *Client) *NodesService {
	return &NodesService{client: c}
}

func (s *NodesService) ListNodes() ([]Node, error) {
	data, err := s.client.Get("/nodes")
	if err != nil {
		return nil, err
	}
	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s *NodesService) GetNode(node string) (json.RawMessage, error) {
	return s.client.Get("/nodes/" + node + "/status")
}
