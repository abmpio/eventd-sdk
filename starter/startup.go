package starter

import (
	"github.com/abmpio/app"
)

func init() {
	app.RegisterOneStartupAction(initEventdClientStartupAction).SetLast()
}
