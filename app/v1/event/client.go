package event

import (
	"encoding/xml"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"net/http"
	"time"
	"fmt"
	"crypto/hmac"
	"crypto/sha1"
	"bytes"
	"io/ioutil"

	"github.com/satori/go.uuid"
	"path"
	"errors"
)

type Config struct {
	// URL for API node callback
	Url string
	// App id
	AppId string
	// App secret
	AppSecret string
}

type Client struct {
	apiUrl *url.URL
	appId string
	appSecret string
}

func NewClient(c Config) (*Client, error) {
	u, err := url.Parse(c.Url)
	if nil != err {
		return nil, fmt.Errorf("Parse URL error: %v", err)
	}

	return &Client{
		apiUrl: u,
		appId: c.AppId,
		appSecret: c.AppSecret,
	}, nil
}

// Post the specified event named 'name' with content 'event' whose type is 'contentType'.
// Content type "application/xml" and "application/json" are supported
func (c *Client) PostEvent(name string, event interface{}, contentType string) error {
	return c.postEvent(name, event, contentType, true)
}

func (c *Client) PostInsecureEvent(name string, event interface{}, contentType string) error {
	return c.postEvent(name, event, contentType, false)
}

func (c *Client) postEvent(name string, event interface{}, contentType string, secure bool) error {
	if name == "" {
		return errors.New("Missing event name")
	}

	u := *c.apiUrl
	u.Path = path.Join(u.Path, name)

	var encoder eventEncoder
	buf := bytes.NewBuffer(nil)
	switch contentType {
	case "application/xml":
		encoder = xml.NewEncoder(buf)
	default:
		contentType = "application/json"
		encoder = json.NewEncoder(buf)
	}

	if err := encoder.Encode(event); nil != err {
		return fmt.Errorf("Encode event error: %v", err)
	}

	if secure {
		u.RawQuery = c.genQuery(u.Query()).Encode()
	}

	resp, err := http.Post(u.String(), contentType, buf)
	if nil != err {
		return fmt.Errorf("Send request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := ioutil.ReadAll(resp.Body)
		var perr *PostErr
		json.Unmarshal(data, &perr)

		if nil == perr {
			return fmt.Errorf("API response error, status code: %d, body: %s", resp.StatusCode, string(data))
		}

		perr.Status = resp.StatusCode
		return perr
	}

	return nil
}

type PostErr struct {
	// Status code
	Status int
	// Error code
	Code int `json:"code"`
	// Error Message
	ErrMsg string `json:"err_msg"`
}

func (e *PostErr) Error() string {
	return fmt.Sprintf("Post event API response error: %d - %s", e.Code, e.ErrMsg)
}

const methodHmacSha1 = "HMAC-SHA1"

const apiMethod = http.MethodPost

type eventEncoder interface {
	Encode(interface{}) error
}

func (c *Client) genQuery(q url.Values) url.Values {
	q.Set("ts", time.Now().UTC().Format(time.RFC3339))
	q.Set("signNonce", uuid.NewV4().String())
	q.Set("signMethod", methodHmacSha1)
	q.Set("signVer", "v1.0")

	plain := fmt.Sprintf("%s&%s&%s", url.QueryEscape(c.appId), apiMethod, q.Encode())
	q.Set("sign", c.sign(plain))

	return q
}

func (c *Client) sign(plain string) string {
	h := hmac.New(sha1.New, []byte(c.appSecret))
	h.Write([]byte(plain))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}