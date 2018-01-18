
# How to use

```
func postEvent() {
	c, err := event.NewClient(event.Config{
		Url: "http://localhost:8080/v1/events",
		AppId: "foo",
		AppSecret: "key3",
	})
	if nil != err {
		log.Fatal(err)
	}

	log.Println(c.PostEvent("dlr", struct {}{}, ""))
}
```

