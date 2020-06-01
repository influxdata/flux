package discord_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/runtime"
)

func TestDiscord(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, `
import "contrib/chobbs/discord"
send = discord.send(webhookToken:"ThisIsAFakeToken",webhookID:"123456789",username:"chobbs",content:"this is fake content!",avatar_url:"%s/somefakeurl.com/pic.png")
send == 204
`)

	if err != nil {
		t.Error("evaluation of discord.send failed: ", err)
	}
	_ = scope
}
