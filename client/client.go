package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abmpio/eventd_sdk/options"
	"github.com/nats-io/nats.go"
)

type EventdClient struct {
	_nc *nats.Conn
}

var (
	_client *EventdClient
)

func NewClient() *EventdClient {
	c := &EventdClient{}
	return c
}

func Client() *EventdClient {
	return _client
}

func SetGlobalClient(c *EventdClient) {
	_client = c
}

func (c *EventdClient) Connect() error {
	// Connect Options.
	opts := []nats.Option{nats.Name("eventd publisher")}

	nkeyFilePath := options.GetOptions().NKeyFile
	// Use Nkey authentication.
	if len(nkeyFilePath) > 0 {
		fullFilePath, err := c.normalizePath(nkeyFilePath)
		if err != nil {
			panic(fmt.Errorf("无效的nkeyFile配置,nkeyFile:%s", nkeyFilePath))
		}
		opt, err := nats.NkeyOptionFromSeed(fullFilePath)
		if err != nil {
			panic(err)
		}
		opts = append(opts, opt)
	}
	natsUrl := fmt.Sprintf("nats://%s:%d", options.GetOptions().Host, options.GetOptions().Port)
	// Connect to eventd
	nc, err := nats.Connect(natsUrl, opts...)
	if err != nil {
		return err
	}
	c._nc = nc
	return nil
}

func (c *EventdClient) normalizePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	curPath, _ := os.Getwd()
	joinedPath := filepath.Join(curPath, "etc", path)
	absolutePath, err := filepath.Abs(joinedPath)
	if err != nil {
		return "", err
	}
	return absolutePath, nil
}

// publish event
func Publish(topic string, v interface{}) error {
	if err := assetClientInit(); err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("序列化v时出现异常,err:%s", err.Error())
	}
	return _client._nc.Publish(topic, data)
}

func Subscribe(topic string, cb nats.MsgHandler) (*nats.Subscription, error) {
	if err := assetClientInit(); err != nil {
		return nil, err
	}
	return _client._nc.Subscribe(topic, cb)
}

func assetClientInit() error {
	if _client == nil || _client._nc == nil {
		return errors.New("请先调用SetGlobalClient函数设置全局client")
	}
	return nil
}
