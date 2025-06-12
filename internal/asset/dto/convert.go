package dto

import (
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// ToBaseAsset 将请求转换为基础资产模型
func (r *BaseRequest) ToBaseAsset(assetType string) *model.BaseAsset {
	tagsBytes, _ := json.Marshal(r.Tags)
	return &model.BaseAsset{
		AssetType:      assetType,
		Name:           r.Name,
		Status:         r.Status,
		ProjectID:      r.ProjectID,
		OrganizationID: r.OrganizationID,
		Tags:           string(tagsBytes),
		CreatedBy:      r.CreatedBy,
		UpdatedBy:      r.UpdatedBy,
	}
}

// ToRequirementAsset 将 CreateRequirementRequest 转为 RequirementAsset
func (r *CreateRequirementRequest) ToRequirementAsset() *model.RequirementAsset {
	stakeholders, _ := json.Marshal(r.Stakeholders)
	acceptanceCriteria, _ := json.Marshal(r.AcceptanceCriteria)
	relatedDocuments, _ := json.Marshal(r.RelatedDocuments)
	return &model.RequirementAsset{
		BusinessValue:      r.BusinessValue,
		Stakeholders:       stakeholders,
		Priority:           r.Priority,
		AcceptanceCriteria: acceptanceCriteria,
		RelatedDocuments:   relatedDocuments,
		Version:            r.Version,
	}
}

// ToDesignDocumentAsset 将 CreateDesignDocumentRequest 转为 DesignDocumentAsset
func (r *CreateDesignDocumentRequest) ToDesignDocumentAsset() *model.DesignDocumentAsset {
	components, _ := json.Marshal(r.Components)
	diagrams, _ := json.Marshal(r.Diagrams)
	dependencies, _ := json.Marshal(r.Dependencies)
	techStack, _ := json.Marshal(r.TechnologyStack)
	return &model.DesignDocumentAsset{
		DesignType:      r.DesignType,
		Components:      components,
		Diagrams:        diagrams,
		Dependencies:    dependencies,
		TechnologyStack: string(techStack),
	}
}

// ToRepositoryAsset 将 CreateRepositoryRequest 转为 RepositoryAsset
func (r *CreateRepositoryRequest) ToRepositoryAsset() *model.RepositoryAsset {
	cicdConfig, _ := json.Marshal(r.CICDConfig)
	return &model.RepositoryAsset{
		RepoURL:        r.RepoURL,
		Branch:         r.Branch,
		LastCommitHash: r.LastCommitHash,
		LastCommitTime: r.LastCommitTime,
		Language:       r.Language,
		CICDConfig:     cicdConfig,
	}
}

// ToUploadedFileAsset 将 CreateUploadedFileRequest 转为 UploadedFileAsset
func (r *CreateUploadedFileRequest) ToUploadedFileAsset() *model.UploadedFileAsset {
	return &model.UploadedFileAsset{
		FilePath:   r.FilePath,
		FileSize:   r.FileSize,
		FileType:   r.FileType,
		Checksum:   r.Checksum,
		PreviewURL: r.PreviewURL,
	}
}

// ToImageAsset 将 CreateImageRequest 转为 ImageAsset
func (r *CreateImageRequest) ToImageAsset() *model.ImageAsset {
	vulnerabilities, _ := json.Marshal(r.Vulnerabilities)
	return &model.ImageAsset{
		RegistryURL:     r.RegistryURL,
		ImageName:       r.ImageName,
		Tag:             r.Tag,
		Digest:          r.Digest,
		Size:            r.Size,
		Vulnerabilities: vulnerabilities,
	}
}

// ToDomainAsset 将 CreateDomainRequest 转为 DomainAsset
func (r *CreateDomainRequest) ToDomainAsset() *model.DomainAsset {
	dnsServers, _ := json.Marshal(r.DNSServers)
	return &model.DomainAsset{
		DomainName:    r.DomainName,
		Registrar:     r.Registrar,
		ExpiryDate:    r.ExpiryDate,
		DNSServers:    string(dnsServers),
		SSLExpiryDate: r.SSLExpiryDate,
	}
}

// ToIPAsset 将 CreateIPRequest 转为 IPAsset
func (r *CreateIPRequest) ToIPAsset() *model.IPAsset {
	return &model.IPAsset{
		IPAddress:   r.IPAddress,
		SubnetMask:  r.SubnetMask,
		Gateway:     r.Gateway,
		DHCPEnabled: r.DHCPEnabled,
		DeviceType:  r.DeviceType,
		MACAddress:  r.MACAddress,
	}
}
