package type_parse

import (
	"encoding/json"
	"fmt"
)

// MapToRawMessage 将 map[string]interface{} 转换为 *json.RawMessage
// 参数: dataMap - 需要转换的 map 数据
// 返回: *json.RawMessage 指针和错误信息
func MapToRawMessage(dataMap map[string]interface{}) (*json.RawMessage, error) {
	// 序列化 map 为 JSON 字节
	rawBytes, err := json.Marshal(dataMap)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	// 转换为 RawMessage 并取地址
	rawJSON := json.RawMessage(rawBytes)
	return &rawJSON, nil
}

// RawMessageToMap 将 *json.RawMessage 转换回 map[string]interface{}
// 参数: rawMsg - 需要解码的 RawMessage 指针
// 返回: map 数据和错误信息
func RawMessageToMap(rawMsg *json.RawMessage) (map[string]interface{}, error) {
	var dataMap map[string]interface{}

	if err := json.Unmarshal(*rawMsg, &dataMap); err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}
	return dataMap, nil
}
