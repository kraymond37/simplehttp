package simplehttp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
		client:   http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}},
	}
}

func MapToUrlValues(content map[string]interface{}) url.Values {
	query := url.Values{}
	for k, v := range content {
		var queryVal string
		switch t := v.(type) {
		case string:
			queryVal = t
		default:
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

func (c *Client) Get(path string, query map[string]interface{}, header http.Header) ([]byte, error) {
	u := c.BuildRequestUrl(path, query)
	return c.sendHttp("GET", u.String(), header, nil)
}

func (c *Client) PostForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("POST", path, params, header)
}

func (c *Client) PostJson(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendJsonRequest("POST", path, params, header)
}

func (c *Client) PutForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("PUT", path, params, header)
}

func (c *Client) PutJson(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendJsonRequest("PUT", path, params, header)
}

func (c *Client) DeleteForm(path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	return c.sendFormRequest("DELETE", path, params, header)
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

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println(string(respBody))
		return nil, fmt.Errorf(resp.Status)
	}

	return respBody, nil
}

func (c *Client) sendFormRequest(verb, path string, params map[string]interface{}, header http.Header) ([]byte, error) {
	values := MapToUrlValues(params)
	data := values.Encode()
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
