package v2

import (
	"github.com/bisonai/v3-bitget-api-sdk/pkg/common"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/utils"
)

type EarnClient struct {
	BitgetRestClient *common.BitgetRestClient
}

func (p *EarnClient) Init(opts ...common.ClientOption) *EarnClient {
	p.BitgetRestClient = new(common.BitgetRestClient).Init(opts...)
	return p
}

func (p *EarnClient) GetProductList(params map[string]string) ([]byte, error) {
	return p.BitgetRestClient.DoGet("/api/v2/earn/savings/product", params)
}

func (p *EarnClient) GetSavingsAssets(params map[string]string) ([]byte, error) {
	return p.BitgetRestClient.DoGet("/api/v2/earn/savings/assets", params)
}

func (p *EarnClient) SubscribeSavings(params map[string]string) ([]byte, error) {
	postBody, jsonErr := utils.ToJson(params)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return p.BitgetRestClient.DoPost("/api/v2/earn/savings/subscribe", postBody)
}

func (p *EarnClient) RedeemSavings(params map[string]string) ([]byte, error) {
	postBody, jsonErr := utils.ToJson(params)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return p.BitgetRestClient.DoPost("/api/v2/earn/savings/redeem", postBody)
}
