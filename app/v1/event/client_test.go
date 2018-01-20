package event

import (
	"testing"
)

func TestClient_PostEvent(t *testing.T) {
	c, err := NewClient(Config{
		Url: "https://d-app.whispir.cc/a/app-027d30049900adf1/v1/events",
		AppId: "027d30049900adf1",
		AppSecret: "pULnwkHww9RuRFTcN9H3E9mAd0gcYBSmzAOqAYSEoBU=",
	})
	if nil != err {
		t.Fatal(err)
	}

	t.Log(c.PostEvent("foo", map[string]string{"to":"haborhuang@whispir.cc", "content":"hello world"}, "application/json"))
}
