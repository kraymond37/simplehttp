package simplehttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Client struct {
	endpoint *url.URL
	client   http.Client
}

func NewClient(endpoint string) *Client {
	s := strings.TrimRight(endpoint, "/")
	u, err := url.Parse(s)
	if err != nil {
		fmt.Println("invalid endpoint")
		return nil
	}

	return &Client{
		endpoint: u,
		client: http.Client{
			Timeout:   20 * time.Second,
			Transport: &http.Transport{DisableKeepAlives: true, Proxy: http.ProxyFromEnvironment},
		},
	}
}

func MapToUrlValues(content map[string]interface{}) url.Values {
	query := url.Values{}
	for k, v := range content {
		var queryVal string
		typeOfValue := reflect.TypeOf(v)
		valueOfValue := reflect.ValueOf(v)
		if typeOfValue.Kind().String() == "string" {
			queryVal = fmt.Sprintf("%s", v)
		} else if typeOfValue.String() == "*string" {
			queryVal = *v.(*string)
		} else if typeOfValue.Kind().String() == "ptr" && typeOfValue.Elem().Kind().String() == "string" {
			queryVal = fmt.Sprintf("%s", valueOfValue.Elem())
		} else {
			j, err := json.Marshal(v)
			if err != nil {
				fmt.Println("marshal value failed", v)
				continue
			}
			queryVal = string(j)
		}
		query.Add(k, queryVal)
	}
	return query
}

func (c *Client) BuildRequestUrl(path string, query map[string]interface{}) *url.URL {
	values := MapToUrlValues(query)
	u := *c.endpoint
	u.Path += path
	u.RawQuery = values.Encode()
	return &u
}

func (c *Client) BuildRequestUrlWithRawQuery(path string, rawQuery string) *url.URL {
	u := *c.endpoint
	u.Path += path
	u.RawQuery = rawQuery
	return &u
}

func (c *Client) Get(path string, query map[string]interface{}, header http.Header) ([]byte, error) {
	u := c.BuildRequestUrl(path, query)
	return c.sendHttp("GET", u.String(), header, nil)
}

// GetString
//
// 某些api对参数顺序有要求, 只能自行构建query字符串
func (c *Client) GetString(path string, rawQuery string, header http.Header) ([]byte, error) {
	u := c.BuildRequestUrlWithRawQuery(path, rawQuery)
	return c.sendHttp("GET", u.String(), header, nil)
}

func (c *Client) PostForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("POST", path, params, header)
}

// PostFormString
//
// 某些api对参数顺序有要求, 只能自行构建body字符串
func (c *Client) PostFormString(path string, body string, header http.Header) ([]byte, error) {
	return c.sendFormString("POST", path, body, header)
}

func (c *Client) PostJson(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendJsonRequest("POST", path, params, header)
}

func (c *Client) PutForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("PUT", path, params, header)
}

// PutFormString
//
// 某些api对参数顺序有要求, 只能自行构建body字符串
func (c *Client) PutFormString(path string, body string, header http.Header) ([]byte, error) {
	return c.sendFormString("PUT", path, body, header)
}

func (c *Client) PutJson(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendJsonRequest("PUT", path, params, header)
}

func (c *Client) DeleteForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("DELETE", path, params, header)
}

// DeleteFormString
//
// 某些api对参数顺序有要求, 只能自行构建body字符串
func (c *Client) DeleteFormString(path string, body string, header http.Header) ([]byte, error) {
	return c.sendFormString("DELETE", path, body, header)
}

func (c *Client) DeleteJson(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendJsonRequest("DELETE", path, params, header)
}

func (c *Client) sendHttp(verb, u string, header http.Header, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(verb, u, body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf(resp.Status)
	}

	return respBody, nil
}

func (c *Client) sendFormRequest(verb, path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	values := MapToUrlValues(params)
	data := values.Encode()
	return c.sendFormString(verb, path, data, header)
}

func (c *Client) sendJsonRequest(verb, path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	data, _ := json.Marshal(params)
	body := strings.NewReader(string(data))
	u := c.BuildRequestUrl(path, nil)
	var h http.Header
	if header != nil {
		h = header
	} else {
		h = http.Header{}
	}
	if body.Len() > 0 {
		h.Set("Content-Type", "application/json")
		h.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	}

	return c.sendHttp(verb, u.String(), h, body)
}

func (c *Client) sendFormString(verb, path string, data string, header http.Header) ([]byte, error) {
	body := strings.NewReader(data)
	u := c.BuildRequestUrl(path, nil)
	var h http.Header
	if header != nil {
		h = header
	} else {
		h = http.Header{}
	}
	if body.Len() > 0 {
		h.Set("Content-Type", "application/x-www-form-urlencoded")
		h.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	}

	return c.sendHttp(verb, u.String(), h, body)
}
