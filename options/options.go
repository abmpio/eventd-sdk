package options

import (
	"fmt"
	"sync"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/configurationx"
	"github.com/go-viper/mapstructure/v2"
)

const (
	ConfigurationKey string = "eventd"
)

var (
	_options EventdOptions
	_once    sync.Once
)

type EventdOptions struct {
	NKeyFile string `json:"nkeyFile"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Disabled bool   `json:"disabled"`
	NKey     string `json:"nkey"`
}

func (o *EventdOptions) normalize() {
	if len(o.Host) <= 0 {
		o.Host = "127.0.0.1"
	}
	if o.Port <= 0 {
		o.Port = 4222
	}
}

func (o *EventdOptions) String() string {
	return fmt.Sprintf("%s:%d",
		o.Host,
		o.Port)
}

func GetOptions() *EventdOptions {
	_once.Do(func() {
		if err := configurationx.GetInstance().UnmarshFromKey(ConfigurationKey, &_options, func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
		}); err != nil {
			err = fmt.Errorf("无效的配置文件,%s", err)
			log.Logger.Error(err.Error())
			panic(err)
		}
		_options.normalize()
	})
	return &_options
}
