package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/eventd-sdk/options"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
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

// check eventd is disabled?
func EventdDisabled() bool {
	return options.GetOptions().Disabled
}

func (c *EventdClient) Connect() error {
	// Connect Options.
	opts := []nats.Option{
		nats.Name("eventd publisher"),
		nats.MaxReconnects(-1),
	}

	eventdOptions := options.GetOptions()
	if len(eventdOptions.NKey) > 0 {
		opt, err := nkeyOptionFromValue(eventdOptions.NKey)
		if err != nil {
			panic(fmt.Errorf("无效的nkey配置,nkey:%s", eventdOptions.NKey))
		}
		opts = append(opts, opt)
	} else {
		nkeyFilePath := eventdOptions.NKeyFile
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
	}
	natsUrl := fmt.Sprintf("nats://%s:%d", options.GetOptions().Host, options.GetOptions().Port)
	// Connect to eventd
	nc, err := nats.Connect(natsUrl, opts...)
	if err != nil {
		return err
	}
	nc.SetErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
		log.Logger.Warn(fmt.Sprintf("eventd的订阅出现异常,topic:%s,err:%s",
			s.Subject, err.Error()))
	})
	nc.SetDisconnectHandler(func(c *nats.Conn) {
		clientId, _ := c.GetClientID()
		log.Logger.Info(fmt.Sprintf("disconnect from eventd server,clientId:%d",
			clientId))
	})
	nc.SetReconnectHandler(func(c *nats.Conn) {
		clientId, _ := c.GetClientID()
		log.Logger.Info(fmt.Sprintf("reconnect from eventd server,clientId:%d",
			clientId))
	})
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

// 重连回调
func OnReconnected(fn func()) {
	if err := assetClientInit(); err != nil {
		return
	}
	_client._nc.SetReconnectHandler(func(c *nats.Conn) {
		fn()
	})
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

// NkeyOptionFromSeed will load an nkey pair from a seed file.
// It will return the NKey Option and will handle
// signing of nonce challenges from the server. It will take
// care to not hold keys in memory and to wipe memory.
func nkeyOptionFromValue(nkey string) (nats.Option, error) {
	kp, err := nkeyPairFromValue(nkey)
	if err != nil {
		return nil, err
	}
	// Wipe our key on exit.
	defer kp.Wipe()

	pub, err := kp.PublicKey()
	if err != nil {
		return nil, err
	}
	if !nkeys.IsValidPublicUserKey(pub) {
		return nil, fmt.Errorf("nats: Not a valid nkey user seed")
	}
	sigCB := func(nonce []byte) ([]byte, error) {
		return sigHandler(nonce, nkey)
	}
	return nats.Nkey(string(pub), sigCB), nil
}

// Sign authentication challenges from the server.
// Do not keep private seed in memory.
func sigHandler(nonce []byte, nkey string) ([]byte, error) {
	kp, err := nkeyPairFromValue(nkey)
	if err != nil {
		return nil, fmt.Errorf("unable to extract key pair from key: %q: %w", nkey, err)
	}
	// Wipe our key on exit.
	defer kp.Wipe()

	sig, _ := kp.Sign(nonce)
	return sig, nil
}

func nkeyPairFromValue(nkey string) (nkeys.KeyPair, error) {
	if nkey == "" {
		return nil, fmt.Errorf("nkey值不能为空")
	}
	contents := []byte(nkey)
	defer wipeSlice(contents)
	return nkeys.ParseDecoratedNKey(contents)
}

// Just wipe slice with 'x', for clearing contents of creds or nkey seed file.
func wipeSlice(buf []byte) {
	for i := range buf {
		buf[i] = 'x'
	}
}
