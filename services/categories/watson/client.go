package watson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jedynykaban/testRestWatsonClient/utils"

	log "github.com/Sirupsen/logrus"
)

// Client is Watson API client (interface) to send REST requests
type Client interface {
	GetFeatures(version string, analysisSources map[string]string, features []string) (map[string]Feature, error)
}

var (
	// ErrorUnableToCreateRequest unified error
	ErrorUnableToCreateRequest = errors.New("Unable to create a request")
	// ErrorUnableToSendRequest unified error
	ErrorUnableToSendRequest = errors.New("Unable to send a request to watson")
	// ErrorRequestRejectedByWatson unified error
	ErrorRequestRejectedByWatson = errors.New("Request rejected by watson")
)

// client is basic implemenation of Watson API client interface
type client struct {
	endpoint string
	hProxy   utils.HttpProxy
	op       utils.HttpProxyOptions
}

// Feature represents unified entity of "scorecard" result obtained from Watson service for particular feature (ex. 'categories' or 'tags')
type Feature struct {
	Score float32 `json:"score"`
	Label string  `json:"label"`
}

type watsonResponse struct {
	Usage       string `json:"usage"`
	RetrivedURL string `json:"retrived_url"`
	Language    string `json:"language"`
	Categories  string `json:"categories"`
}

// NewClient creates new IBM watson client implementing Client interface
func NewClient(endpoint string, hProxy utils.HttpProxy, op utils.HttpProxyOptions) Client {
	return &client{
		endpoint: endpoint,
		hProxy:   hProxy,
		op:       op,
	}
}

func (c *client) GetFeatures(version string, analysisSources map[string]string, features []string) (map[string]Feature, error) {
	result := map[string]Feature{}

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}
	query := url.Query()
	query.Set("version", version)
	for asKey, asValue := range analysisSources {
		query.Set(asKey, asValue)
	}
	query.Set("features", strings.Join(features, ","))
	url.RawQuery = query.Encode()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, ErrorUnableToCreateRequest
	}
	rsp, _, err := c.hProxy.Do(req, c.op)
	if err != nil {
		if len(rsp) > 0 {
			log.WithField("response", rsp).Errorf("Error in watson: %v", err)
		}
		return nil, ErrorUnableToSendRequest
	}

	var watsonResponse watsonResponse
	err = json.Unmarshal(rsp, &watsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal passed response, error: %v", err)
	}

	return result, err
}
