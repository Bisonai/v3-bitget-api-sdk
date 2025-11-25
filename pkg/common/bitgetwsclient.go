package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/bisonai/v3-bitget-api-sdk/pkg/config"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/constants"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/model"
	"github.com/bisonai/v3-bitget-api-sdk/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron"
)

type BitgetBaseWsClient struct {
	NeedLogin                     bool
	Connection                    bool
	LoginStatus                   bool
	Listener                      OnReceive
	ErrorListener                 OnReceive
	Ticker                        *time.Ticker
	SendMutex                     *sync.Mutex
	WebSocketClient               *websocket.Conn
	LastReceivedTime              time.Time
	AllSuribe                     *model.Set
	Signer                        *Signer
	ScribeMap                     map[model.SubscribeReq]OnReceive
	ApiKey, Passphrase, SecretKey string
}

type WsClientOption func(*BitgetBaseWsClient)

func WithWsApiKey(apiKey string) WsClientOption {
	return func(c *BitgetBaseWsClient) {
		c.ApiKey = apiKey
	}
}

func WithWsPassphrase(passphrase string) WsClientOption {
	return func(c *BitgetBaseWsClient) {
		c.Passphrase = passphrase
	}
}

func WithWsSecretKey(secretKey string) WsClientOption {
	return func(c *BitgetBaseWsClient) {
		c.SecretKey = secretKey
	}
}

func (p *BitgetBaseWsClient) Init(opts ...WsClientOption) *BitgetBaseWsClient {
	for _, o := range opts {
		o(p)
	}

	p.Connection = false
	p.AllSuribe = model.NewSet()
	p.Signer = new(Signer).Init(p.SecretKey)
	p.ScribeMap = make(map[model.SubscribeReq]OnReceive)
	p.SendMutex = &sync.Mutex{}
	p.Ticker = time.NewTicker(constants.TimerIntervalSecond * time.Second)
	p.LastReceivedTime = time.Now()

	return p
}

func (p *BitgetBaseWsClient) SetListener(msgListener OnReceive, errorListener OnReceive) {
	p.Listener = msgListener
	p.ErrorListener = errorListener
}

func (p *BitgetBaseWsClient) Connect() {

	p.tickerLoop()
	p.ExecuterPing()
}

func (p *BitgetBaseWsClient) ConnectWebSocket() {
	var err error
	p.WebSocketClient, _, err = websocket.DefaultDialer.Dial(config.WsUrl, nil)
	if err != nil {
		return
	}
	p.Connection = true
}

func (p *BitgetBaseWsClient) Login() {
	timesStamp := utils.TimesStampSec()
	sign := p.Signer.Sign(constants.WsAuthMethod, constants.WsAuthPath, "", timesStamp)
	if constants.RSA == config.SignType {
		sign = p.Signer.SignByRSA(constants.WsAuthMethod, constants.WsAuthPath, "", timesStamp)
	}

	loginReq := model.WsLoginReq{
		ApiKey:     p.ApiKey,
		Passphrase: p.Passphrase,
		Timestamp:  timesStamp,
		Sign:       sign,
	}
	var args []interface{}
	args = append(args, loginReq)

	baseReq := model.WsBaseReq{
		Op:   constants.WsOpLogin,
		Args: args,
	}
	p.SendByType(baseReq)
}

func (p *BitgetBaseWsClient) StartReadLoop() {
	go p.ReadLoop()
}

func (p *BitgetBaseWsClient) ExecuterPing() {
	c := cron.New()
	_ = c.AddFunc("*/15 * * * * *", p.ping)
	c.Start()
}
func (p *BitgetBaseWsClient) ping() {
	p.Send("ping")
}

func (p *BitgetBaseWsClient) SendByType(req model.WsBaseReq) {
	json, _ := utils.ToJson(req)
	p.Send(json)
}

func (p *BitgetBaseWsClient) Send(data string) {
	if p.WebSocketClient == nil {
		return
	}
	p.SendMutex.Lock()
	err := p.WebSocketClient.WriteMessage(websocket.TextMessage, []byte(data))
	p.SendMutex.Unlock()
	if err != nil {
	}
}

func (p *BitgetBaseWsClient) tickerLoop() {
	for {
		select {
		case <-p.Ticker.C:
			elapsedSecond := time.Now().Sub(p.LastReceivedTime).Seconds()

			if elapsedSecond > constants.ReconnectWaitSecond {
				p.disconnectWebSocket()
				p.ConnectWebSocket()
			}
		}
	}
}

func (p *BitgetBaseWsClient) disconnectWebSocket() {
	if p.WebSocketClient == nil {
		return
	}
	err := p.WebSocketClient.Close()
	if err != nil {
		return
	}
}

func (p *BitgetBaseWsClient) ReadLoop() {
	for {

		if p.WebSocketClient == nil {
			//time.Sleep(TimerIntervalSecond * time.Second)
			continue
		}

		_, buf, err := p.WebSocketClient.ReadMessage()
		if err != nil {
			continue
		}
		p.LastReceivedTime = time.Now()
		message := string(buf)

		if message == "pong" {
			continue
		}
		jsonMap := utils.JSONToMap(message)

		v, e := jsonMap["code"]

		if e && int(v.(float64)) != 0 {
			p.ErrorListener(message)
			continue
		}

		v, e = jsonMap["event"]
		if e && v == "login" {
			p.LoginStatus = true
			continue
		}

		v, e = jsonMap["data"]
		if e {
			listener := p.GetListener(jsonMap["arg"])
			listener(message)
			continue
		}
		p.handleMessage(message)
	}

}

func (p *BitgetBaseWsClient) GetListener(argJson interface{}) OnReceive {

	mapData := argJson.(map[string]interface{})

	subscribeReq := model.SubscribeReq{
		InstType: fmt.Sprintf("%v", mapData["instType"]),
		Channel:  fmt.Sprintf("%v", mapData["channel"]),
		InstId:   fmt.Sprintf("%v", mapData["instId"]),
	}

	v, e := p.ScribeMap[subscribeReq]

	if !e {
		return p.Listener
	}
	return v
}

type OnReceive func(message string)

func (p *BitgetBaseWsClient) handleMessage(msg string) {
	fmt.Println("default:" + msg)
}
