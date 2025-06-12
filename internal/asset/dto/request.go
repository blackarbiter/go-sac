package dto

import (
	"time"
)

// BaseRequest 所有资产请求的基类
type BaseRequest struct {
	Name           string   `json:"name" binding:"required"`
	Status         string   `json:"status" binding:"required"`
	ProjectID      uint     `json:"project_id"`
	OrganizationID uint     `json:"organization_id" binding:"required"`
	Tags           []string `json:"tags"`
	CreatedBy      string   `json:"created_by" binding:"required"`
	UpdatedBy      string   `json:"updated_by" binding:"required"`
}

// BaseRequestGetter 用于获取 BaseRequest 字段
//go:generate mockgen -source=request.go -destination=mock_request.go -package=dto
// 方便单元测试

type BaseRequestGetter interface {
	GetBaseRequest() BaseRequest
}

// CreateRequirementRequest 创建需求文档请求
type CreateRequirementRequest struct {
	BaseRequest
	BusinessValue      string   `json:"business_value" binding:"required"`
	Stakeholders       []string `json:"stakeholders" binding:"required"`
	Priority           int      `json:"priority" binding:"required"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	RelatedDocuments   []string `json:"related_documents"`
	Version            string   `json:"version" binding:"required"`
}

func (r *CreateRequirementRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateDesignDocumentRequest 创建设计文档请求
type CreateDesignDocumentRequest struct {
	BaseRequest
	DesignType      string   `json:"design_type" binding:"required"`
	Components      []string `json:"components" binding:"required"`
	Diagrams        []string `json:"diagrams"`
	Dependencies    []string `json:"dependencies"`
	TechnologyStack []string `json:"technology_stack"`
}

func (r *CreateDesignDocumentRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateRepositoryRequest 创建代码仓库请求
type CreateRepositoryRequest struct {
	BaseRequest
	RepoURL        string    `json:"repo_url" binding:"required"`
	Branch         string    `json:"branch" binding:"required"`
	LastCommitHash string    `json:"last_commit_hash"`
	LastCommitTime time.Time `json:"last_commit_time"`
	Language       string    `json:"language" binding:"required"`
	CICDConfig     string    `json:"ci_cd_config"`
}

func (r *CreateRepositoryRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateUploadedFileRequest 创建上传文件请求
type CreateUploadedFileRequest struct {
	BaseRequest
	FilePath   string `json:"file_path" binding:"required"`
	FileSize   int64  `json:"file_size" binding:"required"`
	FileType   string `json:"file_type" binding:"required"`
	Checksum   string `json:"checksum" binding:"required"`
	PreviewURL string `json:"preview_url"`
}

func (r *CreateUploadedFileRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateImageRequest 创建容器镜像请求
type CreateImageRequest struct {
	BaseRequest
	RegistryURL     string   `json:"registry_url" binding:"required"`
	ImageName       string   `json:"image_name" binding:"required"`
	Tag             string   `json:"tag" binding:"required"`
	Digest          string   `json:"digest" binding:"required"`
	Size            int64    `json:"size" binding:"required"`
	Vulnerabilities []string `json:"vulnerabilities"`
}

func (r *CreateImageRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateDomainRequest 创建域名请求
type CreateDomainRequest struct {
	BaseRequest
	DomainName    string    `json:"domain_name" binding:"required"`
	Registrar     string    `json:"registrar" binding:"required"`
	ExpiryDate    time.Time `json:"expiry_date" binding:"required"`
	DNSServers    []string  `json:"dns_servers"`
	SSLExpiryDate time.Time `json:"ssl_expiry_date"`
}

func (r *CreateDomainRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}

// CreateIPRequest 创建IP地址请求
type CreateIPRequest struct {
	BaseRequest
	IPAddress   string `json:"ip_address" binding:"required"`
	SubnetMask  string `json:"subnet_mask"`
	Gateway     string `json:"gateway"`
	DHCPEnabled bool   `json:"dhcp_enabled"`
	DeviceType  string `json:"device_type" binding:"required"`
	MACAddress  string `json:"mac_address"`
}

func (r *CreateIPRequest) GetBaseRequest() BaseRequest {
	return r.BaseRequest
}
