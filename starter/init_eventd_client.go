package starter

import (
	"fmt"
	"time"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/app"

	"github.com/abmpio/eventd-sdk/client"
	"github.com/abmpio/eventd-sdk/options"
)

func initEventdClientStartupAction() app.IStartupAction {
	return app.NewStartupAction(func() {
		if app.HostApplication.SystemConfig().App.IsRunInCli {
			return
		}
		option := options.GetOptions()
		if option.Disabled {
			log.Logger.Warn("eventd.disabled参数为true,将禁用eventd插件")
			return
		}
		client.SetGlobalClient(client.NewClient())
		seconds := 5
		d := time.Second * time.Duration(seconds)
		for {
			log.Logger.Debug("准备初始化eventd...")
			err := client.Client().Connect()
			if err == nil {
				log.Logger.Debug("已成功连接到eventd服务器")
				break
			}
			log.Logger.Warn(fmt.Sprintf("无法连接到eventd服务器,err:%s, %d 后重试...",
				err.Error(),
				seconds))
			time.Sleep(d)
		}
	})
}
