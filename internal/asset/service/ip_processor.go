package service

import (
	"context"
	"fmt"
	"net"

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

// Create 创建IP地址资产
func (p *IPProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	ip, ok := extension.(*model.IPAsset)
	if !ok {
		return nil, fmt.Errorf("invalid IP asset type")
	}

	// 创建IP地址资产
	if err := p.repo.CreateIP(ctx, base, ip); err != nil {
		return nil, err
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新IP地址资产
func (p *IPProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	ip, ok := extension.(*model.IPAsset)
	if !ok {
		return fmt.Errorf("invalid IP asset type")
	}

	// 更新IP地址资产
	return p.repo.UpdateIP(ctx, base, ip)
}

// Get 获取IP地址资产
func (p *IPProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, ip, err := p.repo.GetIP(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, ip, nil
}

// Validate 验证IP地址资产数据
func (p *IPProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证IP地址资产数据
	ip, ok := extension.(*model.IPAsset)
	if !ok {
		return fmt.Errorf("invalid IP asset type")
	}

	// 验证IP地址
	if ip.IPAddress == "" {
		return fmt.Errorf("IP address is required")
	}

	// 验证设备类型
	if ip.DeviceType == "" {
		return fmt.Errorf("device type is required")
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
