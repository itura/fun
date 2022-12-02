package fun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func NewHeaders() Config[string] {
	return Config[string]{}
}

type RestClient struct {
	host   string
	Params *HttpParams
	client *http.Client
}

type HttpParams struct {
	Header Config[string]
	Query  Config[string]
}

func NewHttpParams() *HttpParams {
	return &HttpParams{
		Header: Config[string]{},
		Query:  Config[string]{},
	}
}

func (p *HttpParams) SetQuery(query Config[string]) *HttpParams {
	p.Query = query
	return p
}

func (p *HttpParams) SetHeaders(headers Config[string]) *HttpParams {
	p.Header = headers
	return p
}

func (p *HttpParams) Merge(other *HttpParams) *HttpParams {
	return p.
		SetHeaders(p.Header.SetAll(other.Header)).
		SetQuery(p.Query.SetAll(other.Query))
}

func (p *HttpParams) Apply(req *http.Request) *http.Request {
	for k, v := range p.Header {
		req.Header.Set(k, v)
	}

	for k, v := range p.Query {
		req.URL.Query().Add(k, v)
	}

	return req
}

func NewRestClient(host string) *RestClient {
	return &RestClient{
		host: host,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Params: NewHttpParams().
			SetHeaders(NewConfig[string]().
				Set("Content-Type", "application/json")),
	}
}

func (c *RestClient) SetParams(params *HttpParams) *RestClient {
	c.Params = params
	return c
}

func (c *RestClient) SetClient(client *http.Client) *RestClient {
	c.client = client
	return c
}

type RemoteServiceError struct {
	Status     int
	Payload    JSON
	PayloadRaw string
	Res        *http.Response
	Url        string
}

func (e *RemoteServiceError) Error() string {
	return fmt.Sprintf("status %d from %s", e.Status, e.Url)
}

func (c *RestClient) Get(payload interface{}, path string, params ...*HttpParams) (*http.Response, error) {
	return c.send(payload, http.MethodGet, path, nil, params...)
}

func (c *RestClient) Delete(payload interface{}, path string, params ...*HttpParams) (*http.Response, error) {
	return c.send(payload, http.MethodDelete, path, nil, params...)
}

func (c *RestClient) Post(payload interface{}, path string, body interface{}, params ...*HttpParams) (*http.Response, error) {
	return c.sendBody(payload, http.MethodPost, path, body, params...)
}

func (c *RestClient) Patch(payload interface{}, path string, body interface{}, params ...*HttpParams) (*http.Response, error) {
	return c.sendBody(payload, http.MethodPatch, path, body, params...)
}

func (c *RestClient) Put(payload interface{}, path string, body interface{}, params ...*HttpParams) (*http.Response, error) {
	return c.sendBody(payload, http.MethodPut, path, body, params...)
}

func (c *RestClient) sendBody(payload interface{}, method string, path string, body interface{}, params ...*HttpParams) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.send(payload, method, path, bytes.NewReader(jsonBody), params...)
}

func (c *RestClient) send(payload interface{}, method string, path string, body io.Reader, params ...*HttpParams) (*http.Response, error) {
	url := c.url(path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	mergedParams := NewHttpParams().Merge(c.Params)
	for _, p := range params {
		mergedParams = mergedParams.Merge(p)
	}
	req = mergedParams.Apply(req)

	res, err := c.client.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		e := &RemoteServiceError{
			Status: res.StatusCode,
			Url:    url,
			Res:    res,
		}

		if res.ContentLength != 0 {
			var payload JSON
			err := ParseBody(&payload, res)
			if err != nil {
				bodyBytes, err := io.ReadAll(res.Body)
				if err != nil {
					return res, fmt.Errorf("%v caught during %w", err, e)
				}
				e.PayloadRaw = string(bodyBytes)
			} else {
				e.Payload = payload
			}
		}

		return res, e
	}

	if res.ContentLength != 0 {
		return res, ParseBody(payload, res)
	}

	return res, nil
}

func (c *RestClient) url(path string) string {
	return fmt.Sprintf("%s%s", c.host, path)
}

func ParseBody(message interface{}, res *http.Response) error {
	return json.NewDecoder(res.Body).Decode(message)
}
