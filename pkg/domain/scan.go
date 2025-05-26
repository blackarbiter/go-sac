package domain

import "time"

// ScanResult 表示扫描结果
type ScanResult struct {
	TaskID    string                 `json:"task_id"`    // 任务ID
	ScanType  ScanType               `json:"scan_type"`  // 扫描类型
	AssetID   string                 `json:"asset_id"`   // 资产ID
	AssetType AssetType              `json:"asset_type"` // 资产类型
	Status    string                 `json:"status"`     // 扫描状态：success, failed
	Result    map[string]interface{} `json:"result"`     // 扫描结果
	Error     string                 `json:"error"`      // 错误信息
	Timestamp time.Time              `json:"timestamp"`  // 扫描完成时间
}

// NewScanResult 创建扫描结果
func NewScanResult(taskID string, scanType ScanType, assetID string, assetType AssetType) *ScanResult {
	return &ScanResult{
		TaskID:    taskID,
		ScanType:  scanType,
		AssetID:   assetID,
		AssetType: assetType,
		Timestamp: time.Now(),
	}
}

// SetSuccess 设置成功结果
func (r *ScanResult) SetSuccess(result map[string]interface{}) {
	r.Status = "success"
	r.Result = result
}

// SetFailed 设置失败结果
func (r *ScanResult) SetFailed(err string) {
	r.Status = "failed"
	r.Error = err
}
