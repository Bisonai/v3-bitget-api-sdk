package common

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bisonai/v3-bitget-api-sdk/pkg/config"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/constants"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/utils"
)

type BitgetRestClient struct {
	ApiKey       string
	ApiSecretKey string
	Passphrase   string
	BaseUrl      string
	ProxyUrl     string
	HttpClient   http.Client
	Signer       *Signer
}

type ClientOption func(*BitgetRestClient)

func WithApiKey(apiKey string) ClientOption {
	return func(c *BitgetRestClient) {
		c.ApiKey = apiKey
	}
}

func WithApiSecretKey(apiSecretKey string) ClientOption {
	return func(c *BitgetRestClient) {
		c.ApiSecretKey = apiSecretKey
	}
}

func WithPassphrase(passphrase string) ClientOption {
	return func(c *BitgetRestClient) {
		c.Passphrase = passphrase
	}
}

func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *BitgetRestClient) {
		c.BaseUrl = baseUrl
	}
}

func WithProxyUrl(proxyUrl string) ClientOption {
	return func(c *BitgetRestClient) {
		c.ProxyUrl = proxyUrl
	}
}

func (p *BitgetRestClient) Init(opts ...ClientOption) *BitgetRestClient {
	for _, opt := range opts {
		opt(p)
	}
	p.BaseUrl = config.BaseUrl
	p.Signer = new(Signer).Init(p.ApiSecretKey)
	p.HttpClient = http.Client{
		Timeout: time.Duration(config.TimeoutSecond) * time.Second,
	}

	if p.ProxyUrl != "" {
		proxyUrl, err := url.Parse(p.ProxyUrl)
		if err != nil {
			panic(err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		p.HttpClient.Transport = transport
	}

	return p
}

func (p *BitgetRestClient) DoPost(uri string, params string) ([]byte, error) {
	timesStamp := utils.TimesStamp()
	//body, _ := utils.BuildJsonParams(params)

	sign := p.Signer.Sign(constants.POST, uri, params, timesStamp)
	if constants.RSA == config.SignType {
		sign = p.Signer.SignByRSA(constants.POST, uri, params, timesStamp)
	}
	requestUrl := p.BaseUrl + uri

	buffer := strings.NewReader(params)
	request, err := http.NewRequest(constants.POST, requestUrl, buffer)

	utils.Headers(request, p.ApiKey, timesStamp, sign, p.Passphrase)
	if err != nil {
		return nil, err
	}
	response, err := p.HttpClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (p *BitgetRestClient) DoGet(uri string, params map[string]string) ([]byte, error) {
	timesStamp := utils.TimesStamp()
	body := utils.BuildGetParams(params)

	sign := p.Signer.Sign(constants.GET, uri, body, timesStamp)

	requestUrl := p.BaseUrl + uri + body

	request, err := http.NewRequest(constants.GET, requestUrl, nil)
	if err != nil {
		return nil, err
	}
	utils.Headers(request, p.ApiKey, timesStamp, sign, p.Passphrase)

	response, err := p.HttpClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return res, err
}
