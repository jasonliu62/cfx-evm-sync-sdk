package rpc

import "github.com/openweb3/web3go"

// NewClient 创建一个新的 RPC 客户端
func NewClient(url string) *web3go.Client {
	return web3go.MustNewClient(url)
}
