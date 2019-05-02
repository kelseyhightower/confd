package aac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/log"
)

// Client provides a shell for the env client
type Client struct {
	httpClient       *http.Client
	connectionString connectionString
	watchInterval    time.Duration
	jitterRange      time.Duration
}

// NewAzureAppConfigClient returns a new client
func NewAzureAppConfigClient(connectionString string, requestTimeout time.Duration, watchInterval, jitterRange time.Duration) (*Client, error) {
	httpClient := http.DefaultClient
	httpClient.Timeout = requestTimeout
	rand.Seed(time.Now().UnixNano())

	connString, err := parseConnectionString(connectionString)
	return &Client{
		httpClient:       httpClient,
		connectionString: connString,
		watchInterval:    watchInterval,
		jitterRange:      jitterRange,
	}, err
}

func (c *Client) checkForChange(keys []string, lastIndex uint64) (uint64, error) {

	biggestIndex := lastIndex
	for _, key := range keys {
		validLabel := strings.TrimPrefix(key, "/")
		items, err := c.getValuesByLabel(validLabel)

		if err != nil {
			return lastIndex, err
		}

		for _, item := range items {
			timestamp := uint64(item.LastModified.Unix())
			if timestamp > biggestIndex {
				biggestIndex = timestamp
			}
		}
	}

	return biggestIndex, nil
}

// GetValues queries AAC for values matching the keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	log.Debug("Fetching values for keys: %v", keys)
	val := make(map[string]string)
	for _, key := range keys {
		validLabel := strings.TrimPrefix(key, "/")
		items, err := c.getValuesByLabel(validLabel)
		if err != nil {
			return val, err
		}

		for _, item := range items {
			val[item.Key] = item.Value
		}
	}

	log.Debug("Received values: %v", val)
	return val, nil
}

// WatchPrefix periodically scans for updated timestamps on keys and returns highest timestamp found on change
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// Force first fetch
	if waitIndex == 0 {
		return 1, nil
	}

	ticker := time.NewTicker(c.watchInterval)
	for {
		select {
		case <-ticker.C:
			jitterSleepDuration := time.Nanosecond * time.Duration(rand.Int63n(c.jitterRange.Nanoseconds()))
			log.Debug("Sleeping for jitter duration: %v", jitterSleepDuration)
			time.Sleep(jitterSleepDuration)
			if newIndex, err := c.checkForChange(keys, waitIndex); err == nil && newIndex > waitIndex {
				return newIndex, err
			}

		case <-stopChan:
			return 0, nil
		}
	}
}

func (c *Client) getValuesByLabel(label string) (kvItems []kvItem, err error) {
	path := fmt.Sprintf("/kv?label=%s", label)
	method := "GET"

	headers := getSignedRequestHeaders(method, path, []byte{}, c.connectionString)

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.connectionString.Endpoint, path), nil)
	if err != nil {
		return
	}

	for k, v := range headers {
		req.Header.Set(strings.ToLower(k), v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var items struct {
		Items []kvItem `json:"items"`
	}
	err = dec.Decode(&items)
	if err != nil {
		return
	}

	return items.Items, nil
}

func getSignedRequestHeaders(method, urlPathQuery string, body []byte, connString connectionString) map[string]string {
	verb := strings.ToUpper(method)
	date := time.Now().UTC().Format(http.TimeFormat)

	shaSum := sha256.Sum256(body)
	contentSha256 := base64.StdEncoding.EncodeToString(shaSum[:])

	signedHeaders := "x-ms-date;host;x-ms-content-sha256"

	host := strings.Replace(connString.Endpoint, "https://", "", -1)
	signThis := fmt.Sprintf("%s\n%s\n%s;%s;%s", verb, urlPathQuery, date, host, contentSha256)

	mac := hmac.New(sha256.New, connString.getSecret())
	mac.Write([]byte(signThis))

	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return map[string]string{
		"x-ms-date":           date,
		"x-ms-content-sha256": contentSha256,
		"Authorization":       fmt.Sprintf("HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s", connString.ID, signedHeaders, signature),
	}
}
