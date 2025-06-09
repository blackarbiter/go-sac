package service

import (
	"context"
	"fmt"
	"time"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// DomainProcessor 域名处理器
type DomainProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewDomainProcessor 创建域名处理器
func NewDomainProcessor(repo repository.Repository) *DomainProcessor {
	return &DomainProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建域名资产
func (p *DomainProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	domain, ok := extension.(*model.DomainAsset)
	if !ok {
		return nil, fmt.Errorf("invalid domain asset type")
	}

	// 创建域名资产
	if err := p.repo.CreateDomain(ctx, base, domain); err != nil {
		return nil, err
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新域名资产
func (p *DomainProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	domain, ok := extension.(*model.DomainAsset)
	if !ok {
		return fmt.Errorf("invalid domain asset type")
	}

	// 更新域名资产
	return p.repo.UpdateDomain(ctx, base, domain)
}

// Get 获取域名资产
func (p *DomainProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, domain, err := p.repo.GetDomain(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, domain, nil
}

// Validate 验证域名资产数据
func (p *DomainProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证域名资产数据
	domain, ok := extension.(*model.DomainAsset)
	if !ok {
		return fmt.Errorf("invalid domain asset type")
	}

	// 验证域名
	if domain.DomainName == "" {
		return fmt.Errorf("domain is required")
	}

	// 验证注册商
	if domain.Registrar == "" {
		return fmt.Errorf("registrar is required")
	}

	// 验证到期时间
	if domain.ExpiryDate.IsZero() {
		return fmt.Errorf("expiry date is required")
	}

	return nil
}

// CreateDomainRequest 创建域名请求
type CreateDomainRequest struct {
	Name           string    `json:"name"`
	DomainName     string    `json:"domain_name"`
	Registrar      string    `json:"registrar"`
	ExpiryDate     time.Time `json:"expiry_date"`
	DNSServers     []string  `json:"dns_servers"`
	SSLExpiryDate  time.Time `json:"ssl_expiry_date"`
	CreatedBy      string    `json:"created_by"`
	UpdatedBy      string    `json:"updated_by"`
	ProjectID      uint      `json:"project_id"`
	OrganizationID uint      `json:"organization_id"`
	Tags           []string  `json:"tags"`
}

// UpdateDomainRequest 更新域名请求
type UpdateDomainRequest struct {
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	DomainName    string    `json:"domain_name"`
	Registrar     string    `json:"registrar"`
	ExpiryDate    time.Time `json:"expiry_date"`
	DNSServers    []string  `json:"dns_servers"`
	SSLExpiryDate time.Time `json:"ssl_expiry_date"`
	UpdatedBy     string    `json:"updated_by"`
	Tags          []string  `json:"tags"`
}

// DomainResponse 域名响应
type DomainResponse struct {
	model.BaseAsset
	Extension model.DomainAsset
}
