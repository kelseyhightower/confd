package cfgsvc

import (
	"errors"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	INTERNAL_ERROR   = 500
	BUCKET_NOT_FOUND = 404
)

type BucketResponse struct {
	bucket     *Bucket
	statusCode int
	err        error
}

type ErrorResp struct {
	ErrorType string `json:"type"`
	Message   string `json:"message"`
}

func (this *ErrorResp) Error() string {
	return this.Message
}

type WatchAsync struct {
	bucketName    string
	httpClient    *HttpClient
	dynamicBucket *DynamicBucket
	asyncResp     chan *BucketResponse
}

func (this *WatchAsync) watch() <-chan *BucketResponse {
	go func() {
		resp, err := this.httpClient.get(this.bucketName, LATEST_VERSION, true, this.dynamicBucket.GetVersionAsString())
		if err != nil {
			log.Println("Error making request", err)
			this.asyncResp <- &BucketResponse{bucket: nil, err: err, statusCode: INTERNAL_ERROR}
		} else {
			defer resp.Body.Close()
			this.handleResp(resp)
		}
	}()
	return this.asyncResp
}

func (this *WatchAsync) handleResp(resp *http.Response) {
	if isBucketDeleted(resp) {
		this.asyncResp <- &BucketResponse{bucket: nil, err: errors.New("Bucket is deleted"), statusCode: resp.StatusCode}
	} else {
		this.handleChunkedResp(resp)
	}
}

func (this *WatchAsync) handleChunkedResp(resp *http.Response) {
	httpClient := this.httpClient
	asyncResp := this.asyncResp

	this.dynamicBucket.Connected()
	data, err := httpClient.readResponse(resp)
	if err != nil {
		asyncResp <- &BucketResponse{bucket: nil, err: err, statusCode: 500}
	} else {
		this.createNewBucket(data, resp.StatusCode)
	}
}

func (this *WatchAsync) createNewBucket(data []byte, statusCode int) {
	httpClient := this.httpClient
	asyncResp := this.asyncResp

	newBucket, err := httpClient.newBucket(data)
	if err != nil {
		log.Println("Error while fetching bucket ", err)
		asyncResp <- &BucketResponse{bucket: nil, err: err, statusCode: statusCode}
	} else {
		if newBucket.isValid() {
			asyncResp <- &BucketResponse{bucket: newBucket, err: err, statusCode: statusCode}
		} else {
			this.handleInvalidBucket(data)
		}
	}

}

func (this *WatchAsync) handleInvalidBucket(data []byte) {
	asyncResp := this.asyncResp

	errResp := &ErrorResp{}
	err := ffjson.Unmarshal(data, errResp)
	if err != nil {
		asyncResp <- &BucketResponse{bucket: nil, err: errors.New(string(data)), statusCode: INTERNAL_ERROR}
	} else {
		log.Println("Error parsing bucket from watch", errResp)
		if errResp.ErrorType == DELETED {
			asyncResp <- &BucketResponse{bucket: nil, err: errResp, statusCode: BUCKET_NOT_FOUND}
		} else {
			asyncResp <- &BucketResponse{bucket: nil, err: errResp, statusCode: INTERNAL_ERROR}
		}
	}
}

func isBucketDeleted(resp *http.Response) bool {
	if resp.StatusCode == 404 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading data", err)
		}
		log.Println("Error", string(data))
		return true
	}
	return false
}
