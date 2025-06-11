package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/stretchr/testify/assert"
)

const (
	serverURL = "http://127.0.0.1:8088"
	jwtToken  = "6DNZdF40SaBOOme7V3Ks9cZoj42" // 从之前的测试脚本中获取的 token
)

func TestBatchCreateAssetTasks(t *testing.T) {
	// 测试用例：需求文档
	t.Run("Batch Create Requirement Tasks", func(t *testing.T) {
		requirements := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateRequirementRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "Requirement " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "requirement"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				BusinessValue:      "High business value",
				Stakeholders:       []string{"stakeholder1", "stakeholder2"},
				Priority:           1,
				AcceptanceCriteria: []string{"criteria1", "criteria2"},
				RelatedDocuments:   []string{"doc1", "doc2"},
				Version:            "1.0",
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			requirements[i] = service.CreateAssetTaskRequest{
				AssetID:   "req-" + string(rune('A'+i)),
				AssetType: "Requirement",
				Operation: "create",
				Data:      reqMap,
			}
		}

		// 创建批量请求
		req := service.BatchCreateAssetTaskRequest{
			Tasks: requirements,
		}

		// 发送请求
		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：设计文档
	t.Run("Batch Create Design Document Tasks", func(t *testing.T) {
		designs := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateDesignDocumentRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "Design " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "design"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				DesignType:      "architecture",
				Components:      []string{"component1", "component2"},
				Diagrams:        []string{"diagram1", "diagram2"},
				Dependencies:    []string{"dep1", "dep2"},
				TechnologyStack: []string{"tech1", "tech2"},
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			designs[i] = service.CreateAssetTaskRequest{
				AssetID:   "design-" + string(rune('A'+i)),
				AssetType: "DesignDocument",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: designs,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：代码仓库
	t.Run("Batch Create Repository Tasks", func(t *testing.T) {
		repos := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateRepositoryRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "Repo " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "repo"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				RepoURL:        "https://github.com/test/repo" + string(rune('A'+i)),
				Branch:         "main",
				LastCommitHash: "abc123",
				LastCommitTime: time.Now(),
				Language:       "Go",
				CICDConfig:     "pipeline.yml",
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			repos[i] = service.CreateAssetTaskRequest{
				AssetID:   "repo-" + string(rune('A'+i)),
				AssetType: "Repository",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: repos,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：上传文件
	t.Run("Batch Create Uploaded File Tasks", func(t *testing.T) {
		files := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateUploadedFileRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "File " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "file"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				FilePath:   "/path/to/file" + string(rune('A'+i)),
				FileSize:   1024,
				FileType:   "pdf",
				Checksum:   "sha256:abc123",
				PreviewURL: "https://preview.com/file" + string(rune('A'+i)),
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			files[i] = service.CreateAssetTaskRequest{
				AssetID:   "file-" + string(rune('A'+i)),
				AssetType: "UploadedFile",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: files,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：容器镜像
	t.Run("Batch Create Image Tasks", func(t *testing.T) {
		images := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateImageRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "Image " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "image"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				RegistryURL:     "registry.example.com",
				ImageName:       "test/image" + string(rune('A'+i)),
				Tag:             "latest",
				Digest:          "sha256:abc123",
				Size:            1024 * 1024,
				Vulnerabilities: []string{"CVE-2023-1234"},
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			images[i] = service.CreateAssetTaskRequest{
				AssetID:   "image-" + string(rune('A'+i)),
				AssetType: "Image",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: images,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：域名
	t.Run("Batch Create Domain Tasks", func(t *testing.T) {
		domains := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateDomainRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "Domain " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "domain"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				DomainName:    "example" + string(rune('A'+i)) + ".com",
				Registrar:     "GoDaddy",
				ExpiryDate:    time.Now().AddDate(1, 0, 0),
				DNSServers:    []string{"ns1.example.com", "ns2.example.com"},
				SSLExpiryDate: time.Now().AddDate(1, 0, 0),
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			domains[i] = service.CreateAssetTaskRequest{
				AssetID:   "domain-" + string(rune('A'+i)),
				AssetType: "Domain",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: domains,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})

	// 测试用例：IP地址
	t.Run("Batch Create IP Tasks", func(t *testing.T) {
		ips := make([]service.CreateAssetTaskRequest, 10)
		for i := 0; i < 10; i++ {
			req := dto.CreateIPRequest{
				BaseRequest: dto.BaseRequest{
					Name:           "IP " + string(rune('A'+i)),
					Status:         "active",
					ProjectID:      1,
					OrganizationID: 1,
					Tags:           []string{"test", "ip"},
					CreatedBy:      "test-user",
					UpdatedBy:      "test-user",
				},
				IPAddress:   "192.168.1." + string(rune('1'+i)),
				SubnetMask:  "255.255.255.0",
				Gateway:     "192.168.1.1",
				DHCPEnabled: true,
				DeviceType:  "server",
				MACAddress:  "00:11:22:33:44:55",
			}

			// 将请求转换为 map
			reqMap := make(map[string]interface{})
			reqBytes, _ := json.Marshal(req)
			_ = json.Unmarshal(reqBytes, &reqMap)

			ips[i] = service.CreateAssetTaskRequest{
				AssetID:   "ip-" + string(rune('A'+i)),
				AssetType: "IP",
				Operation: "create",
				Data:      reqMap,
			}
		}

		req := service.BatchCreateAssetTaskRequest{
			Tasks: ips,
		}

		body, _ := json.Marshal(req)
		request, _ := http.NewRequest("POST", serverURL+"/api/v1/tasks/asset/batch", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		response, err := client.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		response.Body.Close()
	})
}
