package dto

import (
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// ToModelRequirementAsset 将 CreateRequirementRequest 转为 RequirementAsset
func ToModelRequirementAsset(req *CreateRequirementRequest) *model.RequirementAsset {
	stakeholders, _ := json.Marshal(req.Stakeholders)
	acceptanceCriteria, _ := json.Marshal(req.AcceptanceCriteria)
	relatedDocuments, _ := json.Marshal(req.RelatedDocuments)
	return &model.RequirementAsset{
		BusinessValue:      req.BusinessValue,
		Stakeholders:       stakeholders,
		Priority:           req.Priority,
		AcceptanceCriteria: acceptanceCriteria,
		RelatedDocuments:   relatedDocuments,
		Version:            req.Version,
	}
}

// ToModelDesignDocumentAsset 将 CreateDesignDocumentRequest 转为 DesignDocumentAsset
func ToModelDesignDocumentAsset(req *CreateDesignDocumentRequest) *model.DesignDocumentAsset {
	components, _ := json.Marshal(req.Components)
	diagrams, _ := json.Marshal(req.Diagrams)
	dependencies, _ := json.Marshal(req.Dependencies)
	techStack, _ := json.Marshal(req.TechnologyStack)
	return &model.DesignDocumentAsset{
		DesignType:      req.DesignType,
		Components:      components,
		Diagrams:        diagrams,
		Dependencies:    dependencies,
		TechnologyStack: string(techStack),
	}
}

// ToModelRepositoryAsset 将 CreateRepositoryRequest 转为 RepositoryAsset
func ToModelRepositoryAsset(req *CreateRepositoryRequest) *model.RepositoryAsset {
	cicdConfig, _ := json.Marshal(req.CICDConfig)
	return &model.RepositoryAsset{
		RepoURL:        req.RepoURL,
		Branch:         req.Branch,
		LastCommitHash: req.LastCommitHash,
		LastCommitTime: req.LastCommitTime,
		Language:       req.Language,
		CICDConfig:     cicdConfig,
	}
}

// ToModelUploadedFileAsset 将 CreateUploadedFileRequest 转为 UploadedFileAsset
func ToModelUploadedFileAsset(req *CreateUploadedFileRequest) *model.UploadedFileAsset {
	return &model.UploadedFileAsset{
		FilePath:   req.FilePath,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		Checksum:   req.Checksum,
		PreviewURL: req.PreviewURL,
	}
}

// ToModelImageAsset 将 CreateImageRequest 转为 ImageAsset
func ToModelImageAsset(req *CreateImageRequest) *model.ImageAsset {
	vulnerabilities, _ := json.Marshal(req.Vulnerabilities)
	return &model.ImageAsset{
		RegistryURL:     req.RegistryURL,
		ImageName:       req.ImageName,
		Tag:             req.Tag,
		Digest:          req.Digest,
		Size:            req.Size,
		Vulnerabilities: vulnerabilities,
	}
}

// ToModelDomainAsset 将 CreateDomainRequest 转为 DomainAsset
func ToModelDomainAsset(req *CreateDomainRequest) *model.DomainAsset {
	dnsServers, _ := json.Marshal(req.DNSServers)
	return &model.DomainAsset{
		DomainName:    req.DomainName,
		Registrar:     req.Registrar,
		ExpiryDate:    req.ExpiryDate,
		DNSServers:    string(dnsServers),
		SSLExpiryDate: req.SSLExpiryDate,
	}
}

// ToModelIPAsset 将 CreateIPRequest 转为 IPAsset
func ToModelIPAsset(req *CreateIPRequest) *model.IPAsset {
	return &model.IPAsset{
		IPAddress:   req.IPAddress,
		SubnetMask:  req.SubnetMask,
		Gateway:     req.Gateway,
		DHCPEnabled: req.DHCPEnabled,
		DeviceType:  req.DeviceType,
		MACAddress:  req.MACAddress,
	}
}
