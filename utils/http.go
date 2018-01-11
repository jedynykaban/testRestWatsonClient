package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/now"
)

func init() {
	// expand supported time formats in jinzhu/now pkg
	var supportedTimeFormats = []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	}
	for _, format := range supportedTimeFormats {
		now.TimeFormats = append(now.TimeFormats, format)
	}
}

const (
	//TODO: Possibly can be moved to the cfg file
	maxIdleConnection = 5
)

type HttpProxy interface {
	Do(rq *http.Request, op HttpProxyOptions) ([]byte, http.Header, error)
	HttpGetter
	HttpHeader
}

type HttpGetter interface {
	// Get wraps HttpProxy.Do method and exposes cleaner interface to the caller.
	Get(url string) ([]byte, error)
}

// Head is a struct containing normalized data
// from a HttpHeader Head request.
type Head struct {
	LastModified time.Time
	ETag         string
}

type HttpHeader interface {
	// Head wraps HttpProxy.Do method and exposes cleaner interface to the caller.
	Head(url string) (Head, error)
}

type HttpGetterHeader interface {
	HttpGetter
	HttpHeader
}

type HttpProxyOptions struct {
	MaxRetries       int
	SleepBeforeRetry time.Duration
}

func DefaultHttpProxyOptions() HttpProxyOptions {
	return HttpProxyOptions{
		MaxRetries:       3,
		SleepBeforeRetry: time.Duration(15) * time.Second,
	}
}

type httpProxy struct {
	client *http.Client
}

var _ HttpProxy = &httpProxy{}

func NewHTTPProxy(timeout time.Duration) *httpProxy {
	return &httpProxy{
		&http.Client{Timeout: timeout, Transport: &http.Transport{MaxIdleConnsPerHost: maxIdleConnection}},
	}
}

func normalizeETag(etag string) string {
	if len(etag) == 0 {
		return ""
	}
	etag = strings.Trim(etag, "\"")
	if strings.HasPrefix(etag, "W/") {
		return "" // don't accept weak ETags
	}
	return etag
}

func (c *httpProxy) Head(url string) (Head, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	var lastMod time.Time
	if err != nil {
		return Head{}, err
	}

	op := DefaultHttpProxyOptions()
	_, hds, err := c.Do(req, op)
	if err != nil {
		return Head{}, err
	}

	lastModStr := hds.Get("Last-Modified")
	etagStr := normalizeETag(hds.Get("ETag"))
	if len(lastModStr) == 0 && len(etagStr) == 0 {
		return Head{}, errors.New("Last-Modified nor ETag HEADER found")
	}

	lastMod, err = now.Parse(lastModStr)
	if err != nil {
		lastMod = time.Time{}
	}

	return Head{
		LastModified: lastMod,
		ETag:         etagStr,
	}, nil
}

func (c *httpProxy) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	op := DefaultHttpProxyOptions()
	ret, _, err := c.Do(req, op)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *httpProxy) Do(rq *http.Request, op HttpProxyOptions) ([]byte, http.Header, error) {
	var retryCounter int
	for {
		nErrf := func(err error, serverRes string, statusCode int) error {
			nErr := err
			if err != nil {
				nErr = errors.New(fmt.Sprintf("%s, server response = %s, http status code = %d", err.Error(), serverRes, statusCode))
			}
			return nErr
		}
		retryCounter++
		rsp, st, hds, err := c.doInternal(rq)
		if retryCounter >= op.MaxRetries {
			log.WithFields(log.Fields{
				"retryCounter": retryCounter,
				"maxRetries":   op.MaxRetries,
				"httpStatus":   st,
				"error":        err,
				"url":          rq.URL.String(),
				"httpMethod":   rq.Method,
			}).Debug("HTTP request permanently rejected. MaxRetries limit reached")
			return rsp, hds, nErrf(err, string(rsp), st)
		}
		if err != nil && shouldRetry(st) {
			time.Sleep(op.SleepBeforeRetry)
			continue
		}
		return rsp, hds, nErrf(err, string(rsp), st)
	}
}

func (c *httpProxy) doInternal(rq *http.Request) ([]byte, int, http.Header, error) {
	var status int
	resp, err := c.client.Do(rq)
	if resp != nil {
		defer resp.Body.Close()
		status = resp.StatusCode
	}

	if err != nil {
		return nil, status, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithField("err", err).Warn("Unable to read body data")
	}

	if status != http.StatusOK && status != http.StatusNoContent && status != http.StatusCreated {
		err = fmt.Errorf("HTTP request failed. HTTP status code returned = %d", status)
	}
	return body, status, resp.Header, err
}

// shouldRetry determines based on HTTP Status Code
// if a request should be retried.
func shouldRetry(st int) bool {
	if st >= 400 && st < 500 {
		return false
	}
	return true
}

// SetHTTPClient swaps internal http client. This method exists solely for testing purpose.
func (c *httpProxy) SetHTTPClient(cl *http.Client) {
	c.client = cl
}

// GetLastToken extracts last non empty token from the url
// i.e. http://www.gizmag.com/lacie-12-big-thunderbolt-hard-drive/42878/
// token = 42878
func GetLastToken(aUrl string) (string, error) {
	parsedURL, err := url.Parse(aUrl)
	if err != nil {
		return "", err
	}

	pathTokens := strings.Split(parsedURL.Path, "/")
	for i := len(pathTokens) - 1; i >= 0; i-- {
		if len(pathTokens[i]) > 0 {
			token := pathTokens[i]
			return token, nil
		}
	}

	return "", errors.New("Unable to find a token in the passed URL.")
}
