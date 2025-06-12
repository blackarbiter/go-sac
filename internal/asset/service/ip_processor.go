package service

import (
	"context"
	"fmt"
	"net"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// IPProcessor IP地址处理器
type IPProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewIPProcessor 创建IP地址处理器
func NewIPProcessor(repo repository.Repository) *IPProcessor {
	return &IPProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建IP资产
func (p *IPProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	var req *model.IPAsset
	switch v := extension.(type) {
	case *model.IPAsset:
		req = v
	case *dto.CreateIPRequest:
		req = v.ToIPAsset()
	default:
		return nil, fmt.Errorf("invalid ip asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateIP(ctx, base, req); err != nil {
		return nil, err
	}
	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新IP资产
func (p *IPProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	var req *model.IPAsset
	switch v := extension.(type) {
	case *model.IPAsset:
		req = v
	case *dto.CreateIPRequest:
		req = v.ToIPAsset()
	default:
		return fmt.Errorf("invalid ip asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateIP(ctx, base, req)
}

// Get 获取IP地址资产
func (p *IPProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, ip, err := p.repo.GetIP(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, ip, nil
}

// Validate 验证IP资产数据
func (p *IPProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.IPAsset
	switch v := extension.(type) {
	case *model.IPAsset:
		req = v
	case *dto.CreateIPRequest:
		req = v.ToIPAsset()
	default:
		return fmt.Errorf("invalid ip asset type")
	}
	if req.IPAddress == "" {
		return fmt.Errorf("ip address is required")
	}
	return nil
}

// CreateIPRequest 创建IP地址请求
type CreateIPRequest struct {
	Name           string   `json:"name"`
	IPAddress      net.IP   `json:"ip_address"`
	SubnetMask     string   `json:"subnet_mask"`
	Gateway        string   `json:"gateway"`
	DHCPEnabled    bool     `json:"dhcp_enabled"`
	DeviceType     string   `json:"device_type"`
	MACAddress     string   `json:"mac_address"`
	CreatedBy      string   `json:"created_by"`
	UpdatedBy      string   `json:"updated_by"`
	ProjectID      uint     `json:"project_id"`
	OrganizationID uint     `json:"organization_id"`
	Tags           []string `json:"tags"`
}

// UpdateIPRequest 更新IP地址请求
type UpdateIPRequest struct {
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	IPAddress   net.IP   `json:"ip_address"`
	SubnetMask  string   `json:"subnet_mask"`
	Gateway     string   `json:"gateway"`
	DHCPEnabled bool     `json:"dhcp_enabled"`
	DeviceType  string   `json:"device_type"`
	MACAddress  string   `json:"mac_address"`
	UpdatedBy   string   `json:"updated_by"`
	Tags        []string `json:"tags"`
}

// IPResponse IP地址响应
type IPResponse struct {
	model.BaseAsset
	Extension model.IPAsset
}
