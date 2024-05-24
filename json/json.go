package json

import (
	"encoding/json"
	"github.com/openweb3/web3go/types"
	"os"
)

// ConvertToJSON 将 Block 结构体实例转换为 JSON 格式的函数
func ConvertToJSON(block *types.Block) ([]byte, error) {
	// 将 Block 结构体实例转换为 JSON 格式的字节切片
	jsonData, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// WriteJSONToFile 将 JSON 数据写入文件的函数
func WriteJSONToFile(jsonData []byte, filename string) error {
	// 将 JSON 数据写入文件
	err := os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}
