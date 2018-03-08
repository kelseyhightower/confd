package google

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/kelseyhightower/confd/log"
)

// Client is a private struct containing internal state for the Google Metadata
// service connection.
type Client struct {
	// host is the hostname or IP of the metadata service.
	host string

	// http is a configured http.Client object for HTTP requests against the
	// metadata service.
	http *http.Client
}

// NewGoogleClient returns a client object to query the Google Metadata
// service.
func NewGoogleClient(host string) (*Client, error) {
	c := new(Client)
	c.host = host
	c.http = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		},
	}

	if c.host == "" {
		return nil, fmt.Errorf("A backend host must be provided")
	}

	return c, nil
}

// getValue returns a value from the metadata service as well as the associated
// ETag using the provided client.  The suffix is appended to
// http://<c.host>/computeMetadata/v1/ and can contain query parameters.
//
// Returns a []byte representing the value, the ETag as a string, and an error
// which is non-nil on any error condition.
func (c *Client) getValue(suffix string) ([]byte, string, error) {
	if suffix[0] == '/' {
		// strip any leading '/' we handle this in building the URL
		suffix = suffix[1:]
	}
	url := "http://" + c.host + "/computeMetadata/v1/" + suffix
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Metadata-Flavor", "Google")
	log.Debug("Google Metadata: Fetching %s", url)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, "0", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, "0", fmt.Errorf("metadata key not found: %s", url)
	}
	if res.StatusCode != 200 {
		return nil, "0", fmt.Errorf("status code %d trying to fetch %s", res.StatusCode, url)
	}
	all, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "0", err
	}

	return all, res.Header.Get("Etag"), nil
}

// merge will return the union of maps a and b.  Keys in b will overwrite
// keys in a if there are duplicates.  A reference to a is returned.
func merge(a, b map[string]string) map[string]string {
	for k, v := range b {
		a[k] = v
	}

	return a
}

// formatResult is a helper function to handle building of a map from a varying
// typed data structure.  The only possible error returned is a JSON
// marshalling error which should not occur as the data was just unmarshaled.
func formatResult(key string, data interface{}) (map[string]string, error) {
	ret := make(map[string]string)

	switch value := data.(type) {
	case string:
		ret[key] = value
	case map[string]interface{}:
		for k, v := range value {
			r, err := formatResult(path.Join(key, k), v)
			if err != nil {
				return nil, err
			}
			ret = merge(ret, r)
		}
	default:
		// Probably lists like for instance/tags -- return encoded JSON
		blob, err := json.Marshal(data)
		if err != nil {
			// This shouldn't happen as we just Unmarshed() this data
			log.Debug("Error marshaling JSON data while formatting results")
			return nil, err
		}
		ret[key] = string(blob)
	}

	return ret, nil
}

// GetValues returns a map of the provided keys to their associated values in
// the Google Metadata service.  Recursive lookups are done by default.  For
// keys that map to a simple file or string value a simple string is returned.
// For keys that reference a directory entry or other complex data type a
// string containing raw JSON encoded data is returned.  For a directory this
// will be a map of sub-keys to values or other maps.  For the instance/tags
// key this will be a JSON list of tags.  On error or key not found a non-nil
// error is returned with a partially populated map.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	var data interface{}
	ret := make(map[string]string)

	for _, i := range keys {
		log.Debug("Google Metadata: Getting: %s", i)
		blob, _, err := c.getValue(i + "?alt=json&recursive=true")
		if err != nil {
			log.Info("Google Metadata: Error fetching data: %s", err.Error())
			return ret, err
		}

		// JSON unmarshal the data so we can make some decisions based on it.
		err = json.Unmarshal(blob, &data)
		if err != nil {
			log.Info("Google Metadata: Error parsing JSON: %s", err.Error())
			return ret, err
		}

		hash, err := formatResult(i, data)
		if err != nil {
			return ret, err
		}

		ret = merge(ret, hash)
	}

	return ret, nil
}

// WatchPrefix watches a set of Google Metadata rooted at the prefix string
// for changes.  The slice of keys is ignored for simplicity.  The waitIndex
// is the uint64 encoded ETag number expressing the hash of the metadata
// object and must be 0 on the first call.  The stopChan channel causes
// this function to cancel and return the last known waitIndex.
//
// When changes are detected this function returns the uint64 version of the
// ETag and a nil error object.  On error the last waitIndex is returned and
// err is non-nil.
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	ETag := fmt.Sprintf("%x", waitIndex)
	for waitIndex != 0 && len(ETag) < 16 {
		// padding
		ETag = "0" + ETag
	}
	suffix := prefix +
		"?alt=json&recursive=true&wait_for_change=true&last_etag=" + ETag

	for {
		// Check for the stop signal.  We could do a context object here
		// to make this more immediate, but that adds complexity that is
		// not otherwise found in this codebase.
		select {
		case <-stopChan:
			return waitIndex, nil
		default:
		}

		log.Debug("Google Metadata Watch: Fetching %s with etag different from %s",
			prefix, ETag)
		_, newETag, err := c.getValue(suffix)
		if err != nil || newETag == "" {
			// All errors are treated as retry-able by the calling function with
			// a 2 second sleep.
			log.Info("Google Metadata Watch: %s", err.Error())
			return waitIndex, err
		}

		if newETag == ETag {
			// HTTP call returned with no changes, loop
			continue
		}

		// We have a new ETag -- data under the prefix has changed
		// Google's ETags are a 64 bit hash/checksum value, and I'm assuming that
		// will continue to be the case.  It fits nicely into the waitIndex
		// variable confd uses to determine metadata changes.
		uetag, err := strconv.ParseUint(newETag, 16, 64)
		if err != nil {
			log.Info("Google Metadata Watch: ETag parsing error: %s",
				err.Error())
			return waitIndex, err
		}

		return uetag, nil
	}

	// This is never executed
	return 0, nil
}
