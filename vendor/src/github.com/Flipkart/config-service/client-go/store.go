package cfgsvc

import (
	"errors"
	"log"
    "net"
    "os"
	"net/http"
	"strconv"
	"net/url"
	"github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	"time"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
)


// HttpClient is used to provide abstractions such as "bucket" and "keys" over
// low-level HTTP API such as Get and Watch
type HttpClient struct {
	instance *http.Client
	url string
	ipv4 string
	hostname string
}

// NewHttpClient is the constructor for the bucket API implementation of HttpClient.
func NewHttpClient(client *http.Client, url string) (*HttpClient, error) {

    // get hostname
	hostname, err := os.Hostname()
	if (err != nil) {
		return nil, err
	}

    // get ipv4
    var hostIP string
    interfaces, _ := net.Interfaces()
    for _, inter := range interfaces {
        if addrs, err := inter.Addrs(); err == nil {
            for _, addr := range addrs {
                switch ip := addr.(type) {
                case *net.IPNet:
                    if ip.IP.DefaultMask() != nil && !ip.IP.IsLoopback() {
                        hostIP = ip.IP.To4().String()
                    }
                }
            }
        }
    }

    // create instance
	return &HttpClient{instance: client, url: url, ipv4: hostIP, hostname: hostname}, nil

}

const(
	BUCKET_PATH  = "/v1/buckets/"
	INITIAL_VERSION = "0"
	DELETED = "DELETED"
	NOT_MODIFIED = "NOT_MODIFIED"
)

//getBucketURL builds URL to be used by the HTTP client
func (this *HttpClient) getBucketURL(name string, version int, watch bool) string {
	urlBuilder, err := url.Parse(this.url + BUCKET_PATH + name)
	if err != nil {
		log.Fatal(err)
	}
	query := urlBuilder.Query()

	query.Set("watch", strconv.FormatBool(watch))
	if version >= 0 {
		query.Set("version", strconv.Itoa(version))
	}

	urlBuilder.RawQuery = query.Encode()
	return urlBuilder.String()
}

// gets bucket instance given bucket name and other details
func (this *HttpClient) get(name string, version int, watch bool, sourceVersion string) (*http.Response, error) {

	log.Println("Getting bucket: ", name)

	req, err := http.NewRequest("GET", this.getBucketURL(name, version, watch), nil)
	if err != nil {
		log.Println("Error creating new request", err)
		return nil, err
	}

	req.Header.Add("X-Config-Bucket-Version", sourceVersion)

    // identity headers
    req.Header.Add("X-Client-IPv4", this.ipv4)
    req.Header.Add("X-Client-Hostname", this.hostname)

	resp, err := this.instance.Do(req)
	if err != nil {
		log.Println("Error making request", err)
		return nil, err
	} else {
		return resp, nil
	}
}

//Fetch Bucket from the config service
func (this *HttpClient) GetBucket(name string, version int) (*Bucket, error) {
	// fetch data
	resp, err := this.get(name, version, false, INITIAL_VERSION)
	if err != nil {
		log.Println("Error fetching bucket ", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errResp := &ErrorResp{}
		err := ffjson.NewDecoder().DecodeReader(resp.Body, errResp)
		if err != nil {
			log.Println("Error reading the response Body")
		}
		log.Println("Error fetching bucket: ", errResp)
		return nil, errors.New(errResp.Error())
	}

	// read response
	data, err := this.readResp(resp)
	if err != nil {
		log.Println("Error reading resp ", err.Error())
		return nil, err
	}

	// create and return bucket
	bucket, err := this.newBucket(data)
	if err != nil {
		log.Println("Error creating bucket ", err.Error())
		return nil, err
	}

	return bucket, nil
}

func (this *HttpClient) readResp(resp *http.Response) ([]byte, error) {
	return ioutil.ReadAll(resp.Body)
}

// newBucket creates a bucket from JSON data
func (this *HttpClient) newBucket(data []byte) (*Bucket, error) {
	log.Println("Extracting keys from the response body")

	bucket := &Bucket{}

	err := ffjson.Unmarshal(data, bucket)
	if err != nil {
		return nil, errors.New("Error decoding bucket")
	}

	log.Println("Fetched bucket ", bucket)

	// fetch and decode keys
	return bucket, nil
}

// WatchBucket sets up a watch on a bucket and sends appropriate events to the listener
func (this *HttpClient) WatchBucket(name string, cache *lru.Cache, dynamicBucket *DynamicBucket){
	backOff :=  &backoff.Backoff{
		Min:    1 * time.Second,
		Max: 50 * time.Second,
		Jitter: true,
	}
	for {
		log.Println("Setting watch on bucket: ", name)
		watchAsync := WatchAsync{
			bucketName: name,
			dynamicBucket: dynamicBucket,
			asyncResp: make(chan *BucketResponse),
			httpClient: this,
		}

		select {
		case bucketResp := <- watchAsync.watch():

			if bucketResp.err != nil && bucketResp.statusCode == 404 {
				log.Println("Stopping watch on bucket: ", name)
				dynamicBucket.DeleteBucket()
				cache.Remove(name)
				return
			}

		    if bucketResp.err != nil && bucketResp.statusCode == 304 {
				log.Println("Watch timed out fetching bucket: ", bucketResp.err)
				dynamicBucket.Disconnected(bucketResp.err)
				backOff.Reset()
				continue;
			}

			if bucketResp.err != nil {
				log.Println("Error fetching bucket: ", bucketResp.err)
				dynamicBucket.Disconnected(bucketResp.err)
				time.Sleep(backOff.Duration())
				continue;
			}

		    backOff.Reset()
			dynamicBucket.updateBucket(bucketResp.bucket)

		case <- dynamicBucket.isShutdown():
			log.Println("Stopping watch on bucket: ", name)
			return

		}
	}
}
