package circleci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultBaseURL = "https://www.circleci.com/api/v3/"

// Client is the primary type that implements an interface to the circleci.com API.
type Client struct {
	baseURL string
	token   string
	client  *http.Client
	logger  Logger
}

// NewClient initializes a new Client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL: defaultBaseURL,
		logger:  &WrapLogrus{logrus.New()},
	}
	for _, o := range opts {
		o(c)
	}
	if c.client == nil {
		c.client = http.DefaultClient
	}
	return c, nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

func (c *Client) get(pattern string, args ...interface{}) ([]byte, error) {
	return c.do("GET", nil, pattern, args...)
}

func (c *Client) post(payload interface{}, pattern string, args ...interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}
	c.logger.WithField("fn", "post").Debugln(buf.String())
	return c.do("POST", buf, pattern, args...)
}

func (c *Client) do(method string, body io.Reader, pattern string, args ...interface{}) ([]byte, error) {
	path := c.url(fmt.Sprintf(pattern, args...))
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	req.Header.Set("cookie", fmt.Sprintf("token=%v", c.token))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()
	logger := c.logger.WithField("method", method).WithField("path", path).WithField("status_code", resp.StatusCode)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warnln("error reading body")
		return nil, err
	}
	logger.WithField("body", string(buf)).Debugln("api call finished")
	if resp.StatusCode != http.StatusOK {
		return buf, &Error{
			URL:        path,
			StatusCode: resp.StatusCode,
			Body:       string(buf),
		}
	}
	return buf, nil
}
