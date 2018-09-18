package circleci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/russolsen/transit"
	"github.com/sirupsen/logrus"
	"github.com/tmc/circleci/circletypes"
	"github.com/tmc/transitutils"
)

const defaultBaseURL = "https://circleci.com/query-api"

// Client is the primary type that implements an interface to the circleci.com API.
type Client struct {
	baseURL      string
	token        string
	sessionToken string
	client       *http.Client
	logger       Logger
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

func (c *Client) get(ctx context.Context, pattern string, args ...interface{}) ([]byte, error) {
	return c.do(ctx, "GET", nil, pattern, args...)
}

func (c *Client) post(ctx context.Context, payload io.Reader, pattern string, args ...interface{}) ([]byte, error) {
	return c.do(ctx, "POST", payload, pattern, args...)
}

func (c *Client) do(ctx context.Context, method string, body io.Reader, pattern string, args ...interface{}) ([]byte, error) {
	path := c.url(fmt.Sprintf(pattern, args...))
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	req = req.WithContext(ctx)
	//req.Header.Add("cookie", fmt.Sprintf("token=%v", c.token))
	req.Header.Add("Cookie", fmt.Sprintf("ring-session=%v", c.sessionToken))
	req.Header.Add("Accept", "application/transit+json; charset=UTF-8")
	if body != nil {
		req.Header.Set("Content-Type", "application/transit+json")
	}
	/*
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
			return nil, errors.Wrap(err, "dumping request")
		}
		fmt.Println(string(requestDump))
	*/
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

// GetWorkflow returns a workflow given a workflow id. Must be authenticated with a session token.
func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (*circletypes.Workflow, error) {
	payload := strings.NewReader(fmt.Sprintf(`["^ ","~:type","~:get-workflow-status","~:params",["^ ","~:run/id","~u%v"]]`, workflowID))

	buf, err := c.post(ctx, payload, "")
	if err != nil {
		return nil, errors.Wrap(err, "issue posting")
	}
	value, err := transit.NewDecoder(bytes.NewReader(buf)).Decode()
	if err != nil {
		return nil, errors.Wrap(err, "issue decoding transit")
	}
	v, err := transitutils.ToGo(value)
	if err != nil {
		return nil, errors.Wrap(err, "issue converting transit to go")
	}
	j, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "issue marshaling")
	}
	result := &circletypes.Workflow{}
	err = json.Unmarshal(j, result)
	return result, errors.Wrap(err, "issue unmarshaling")
}
