package pve

import (
	"context"
	"encoding/json"
)

// Storage 表示 PVE 存储池
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Total   int64  `json:"total"`
	Used    int64  `json:"used"`
	Avail   int64  `json:"avail"`
	Active  int    `json:"active"`
}

// StorageService 处理存储池相关业务逻辑
type StorageService struct {
	client *Client
}

func NewStorageService(c *Client) *StorageService {
	return &StorageService{client: c}
}

func (s *StorageService) ListStorage(ctx context.Context) ([]Storage, error) {
	data, err := s.client.Get(ctx, "/storage")
	if err != nil {
		return nil, err
	}
	var storages []Storage
	if err := json.Unmarshal(data, &storages); err != nil {
		return nil, err
	}
	return storages, nil
}

// ensure json import used
var _ = json.RawMessage{}
