package intriniorealtime

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	cQUODDRealtimeTokenURL = "https://api.intrinio.com/token?type=QUODD"
	cQUODDWebsocketURL     = "wss://www5.quodd.com/websocket/webStreamer/intrinio"

	cIEXRealtimeTokenURL = "https://realtime.intrinio.com/auth"
	cIEXWebsocketURL     = "wss://realtime.intrinio.com/socket/websocket"
)

type provider string

const (
	// IEX provider
	IEX provider = "iex"
	// QUODD provider
	QUODD provider = "quodd"
)

// Client Overview
type Client struct {
	DebugMode bool

	username string
	password string
	provider provider

	token          string
	ws             *websocket.Conn
	channels       map[string]bool
	joinedChannels map[string]bool

	quoteHander  func(quote map[string]interface{})
	errorHandler func(err error)

	sending       chan struct{}
	breakHartbeat chan struct{}
	breakSender   chan struct{}
	q             chan map[string]interface{}
}

// New Overview
func New(username, password string, provider provider) *Client {
	return &Client{
		username:       username,
		password:       password,
		provider:       provider,
		DebugMode:      false,
		channels:       make(map[string]bool),
		joinedChannels: make(map[string]bool),
		sending:        make(chan struct{}, 1),
		breakHartbeat:  make(chan struct{}, 1),
		breakSender:    make(chan struct{}, 1),
		q:              make(chan map[string]interface{}),
	}
}

// Connect Overview
func (cli *Client) Connect() error {
	cli.debug("%s\n", "Websocket connecting...")
	if err := cli.refreshToken(); err != nil {
		return err
	}
	return cli.refreshWebsocket()
}

// Disconnect Overview
func (cli *Client) Disconnect() error {
	if cli.connected() == false {
		return nil
	}

	cli.onClosing()

	close(cli.breakHartbeat)
	close(cli.breakSender)
	<-cli.sending

	err := cli.ws.Close()
	if err != nil {
		cli.onCloseFailed()
		return err
	}
	cli.ws = nil
	time.Sleep(time.Second)
	cli.onClosed()
	return err
}

// Join Overview
func (cli *Client) Join(channels ...string) {
	for _, channel := range channels {
		c := strings.TrimSpace(channel)
		if _, ok := cli.channels[c]; !ok {
			cli.channels[c] = true
		}
	}
	cli.refreshChannels()
}

// Leave Overview
func (cli *Client) Leave(channels ...string) {
	for _, channel := range channels {
		fmt.Println(channels)
		fmt.Println(cli.channels)
		delete(cli.channels, strings.TrimSpace(channel))
		fmt.Println(cli.channels)
	}
	cli.refreshChannels()
}

// LeaveAll Overview
func (cli *Client) LeaveAll() {
	cli.channels = make(map[string]bool)
	cli.refreshChannels()
}

func (cli *Client) refreshToken() error {
	req, err := http.NewRequest("GET", makeAuthURL(cli.provider), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(cli.username, cli.password)
	client := &http.Client{Timeout: time.Duration(10) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", "Auth failed.")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	cli.token = string(b)
	return nil
}

func (cli *Client) refreshWebsocket() error {
	if cli.connected() {
		cli.Disconnect()
	}

	c, _, err := websocket.DefaultDialer.Dial(makeSoketURL(cli.provider, cli.token), nil)
	if err != nil {
		return err
	}
	cli.ws = c
	cli.onConnected()
	return nil
}

func (cli *Client) refreshChannels() {
	if cli.connected() == false {
		return
	}
	for k := range cli.channels {
		if _, ok := cli.joinedChannels[k]; !ok {
			cli.q <- makeJoinMessage(cli.provider, k)
		}
	}
	for k := range cli.joinedChannels {
		if _, ok := cli.channels[k]; !ok {
			cli.q <- makeLeaveMessage(cli.provider, k)
		}
	}
	cli.joinedChannels = make(map[string]bool)
	for k := range cli.channels {
		cli.joinedChannels[k] = true
	}
}

func (cli *Client) startReceiver() {
	go func() {
		for {
			var ret map[string]interface{}
			err := cli.ws.ReadJSON(&ret)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					cli.onError(err)
				}
				return
			}
			cli.onQuote(ret)
		}
	}()
}

func (cli *Client) startSender() {
	go func() {
		for {
			select {
			case data := <-cli.q:
				cli.debug("send data = %v\n", data)
				if err := cli.ws.WriteJSON(data); err != nil {
					cli.onError(err)
				}
			case <-cli.breakSender:
				for 0 < len(cli.q) {
					cli.debug("Quit sender! queue count = %d\n", len(cli.q))
					time.Sleep(100 * time.Millisecond)
				}
				close(cli.sending)
				return
			}
		}
	}()
}

func (cli *Client) heartbeat() {
	go func() {
		hearbeatTime := time.NewTicker(3 * time.Second)
		defer hearbeatTime.Stop()
		for {
			select {
			case <-hearbeatTime.C:
				cli.q <- makeHeartbeatMessage(cli.provider)
			case <-cli.breakHartbeat:
				return
			}
		}
	}()
}

func (cli *Client) connected() bool {
	return cli.ws != nil
}

func (cli *Client) debug(format string, a ...interface{}) {
	if cli.DebugMode == false {
		return
	}
	fmt.Printf(format, a...)
}

func (cli *Client) onConnected() {
	cli.debug("%s\n", "Websocket connected")
	cli.startReceiver()
	cli.startSender()
	cli.heartbeat()
}

func (cli *Client) onClosing() {
	cli.debug("%s\n", "Websocket closing")
}
func (cli *Client) onCloseFailed() {
	cli.debug("%s\n", "Websocket failed close")
}

func (cli *Client) onClosed() {
	cli.debug("%s\n", "Websocket closed")
}

// OnQuote Overview
func (cli *Client) OnQuote(f func(map[string]interface{})) {
	cli.quoteHander = f
}

func (cli *Client) onQuote(a map[string]interface{}) {
	cli.debug("%v\n", a)
	if cli.quoteHander != nil {
		cli.quoteHander(a)
	}
}

// OnError Overview
func (cli *Client) OnError(f func(err error)) {
	cli.errorHandler = f
}

func (cli *Client) onError(err error) {
	cli.debug("IntrinioRealtime | Websocket error: %v\n", err)
	if cli.errorHandler != nil {
		cli.errorHandler(err)
	}
}

func makeAuthURL(provider provider) string {
	switch provider {
	case IEX:
		return cIEXRealtimeTokenURL
	case QUODD:
		return cQUODDRealtimeTokenURL
	default:
		panic("A value that does not exist was specified.")
	}
}

func makeSoketURL(provider provider, token string) string {
	switch provider {
	case IEX:
		return fmt.Sprintf("%s?vsn=1.0.0&token=%s", cIEXWebsocketURL, token)
	case QUODD:
		return fmt.Sprintf("%s/%s", cQUODDWebsocketURL, token)
	default:
		panic("A value that does not exist was specified.")
	}
}

func makeJoinMessage(provider provider, channel string) map[string]interface{} {
	if provider == IEX {
		return map[string]interface{}{
			"topic":   parseTopic(channel),
			"event":   "phx_join",
			"payload": map[string]interface{}{},
			"ref":     nil,
		}
	} else if provider == QUODD {
		return map[string]interface{}{
			"event": "subscribe",
			"data": map[string]string{
				"ticker": channel,
				"action": "subscribe",
			},
		}
	} else {
		panic("A value that does not exist was specified.")
	}
}

func makeLeaveMessage(provider provider, channel string) map[string]interface{} {
	if provider == IEX {
		return map[string]interface{}{
			"topic":   parseTopic(channel),
			"event":   "phx_leave",
			"payload": map[string]interface{}{},
			"ref":     nil,
		}
	} else if provider == QUODD {
		return map[string]interface{}{
			"event": "unsubscribe",
			"data": map[string]string{
				"ticker": channel,
				"action": "unsubscribe",
			},
		}
	} else {
		panic("A value that does not exist was specified.")
	}
}

func makeHeartbeatMessage(provider provider) map[string]interface{} {
	if provider == IEX {
		return map[string]interface{}{
			"topic":   "phoenix",
			"event":   "heartbeat",
			"payload": map[string]interface{}{},
			"ref":     nil,
		}
	} else if provider == QUODD {
		return map[string]interface{}{
			"event": "heartbeat",
			"data": map[string]interface{}{
				"action": "heartbeat",
				"ticker": time.Now().Unix(),
			},
		}
	} else {
		panic("A value that does not exist was specified.")
	}
}

func parseTopic(channel string) string {
	if channel == "$lobby" {
		return "iex:lobby"
	} else if channel == "$lobby_last_price" {
		return "iex:lobby:last_price"
	}
	return "iex:securities:" + channel
}
