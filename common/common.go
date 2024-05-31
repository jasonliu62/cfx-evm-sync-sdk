package common

import "github.com/openweb3/web3go"

type GetFunc func(*web3go.Client, interface{}) (interface{}, error)

type DataWrap struct {
	Error error
	Value interface{}
}
