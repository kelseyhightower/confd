package config_service

import (
	"strings"
	cfgsvc "github.com/Flipkart/config-service/client-go"
	"github.com/kelseyhightower/confd/log"
	"errors"
	"reflect"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
)

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *cfgsvc.ConfigServiceClient
}

type BucketListener struct{
	watchResp chan *watchResponse
	currentIndex uint64
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

func (this *BucketListener) Connected(bucketName string) {
	log.Info("Connected! " + bucketName)
}

func (this *BucketListener) Disconnected(bucketName string, err error) {
	log.Info("Disconnected! " + bucketName)
	this.watchResp <- &watchResponse{waitIndex:this.currentIndex, err: err}
}

func (this *BucketListener) Deleted(bucketName string) {
	log.Info("deleted " + bucketName)
	this.watchResp <- &watchResponse{waitIndex: 0, err: errors.New(bucketName + " was deleted")}
}

func (this *BucketListener) Updated(oldBucket *cfgsvc.Bucket, newBucket *cfgsvc.Bucket) {
	this.watchResp <- &watchResponse{waitIndex:this.currentIndex+1, err: nil}
}


func NewConfigClient(machines []string) (*Client, error) {
	c, err := cfgsvc.NewConfigServiceClient(machines[0], 50) //*10)
	if err != nil {
		panic(err)
	}
	return &Client{c}, nil
}


func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range keys {
		bucketsKey := strings.Split(strings.TrimPrefix(v, "/"), "/")
		buckets := strings.Split(bucketsKey[0], ",")
		key := bucketsKey[1]

		dynamicBuckets, err := c.getDynamicBuckets(buckets)
		if err != nil {
			return vars, err
		}


		for _, dynamicBucket := range dynamicBuckets {
			val := dynamicBucket.GetKeys()[key]
			if val == nil {
				continue;
			}
			valType := reflect.TypeOf(val).Kind()
			if valType == reflect.Slice {
				data, err := ffjson.Marshal(val)
				if err != nil {
				    log.Error("Failed decoding from JSON")
				} else {
					vars[key] = string(data[:])
				}
			} else {
				switch val.(type) {
					case int,int64:
					vars[key] = strconv.FormatInt(val.(int64), 64)
					case string:
					vars[key] = val.(string)
					case bool:
					vars[key] = strconv.FormatBool(val.(bool))
					case float32,float64:
					vars[key] = strconv.FormatFloat(val.(float64), 'f', -1, 64)
				}
			}
		}

	}
	return vars, nil
}

func (c *Client) getDynamicBuckets(buckets []string) ([]*cfgsvc.DynamicBucket, error) {
	var dynamicBuckets []*cfgsvc.DynamicBucket
	for _, bucket := range buckets {
		bucketName := strings.TrimSpace(bucket)
		dynamicBucket, err := c.client.GetDynamicBucket(bucketName)
		if err != nil {
			return dynamicBuckets, err
		}
		dynamicBuckets = append(dynamicBuckets, dynamicBucket)
	}
	return dynamicBuckets, nil
}

func setupDynamicBucketListeners(dynamicBuckets []*cfgsvc.DynamicBucket, bucketListener *BucketListener) {
	for _, dynamicBucket := range dynamicBuckets {
		dynamicBucket.AddListeners(bucketListener)
	}
}

func removeDynamicBucketListeners(dynamicBuckets []*cfgsvc.DynamicBucket, bucketListener *BucketListener) {
	for _, dynamicBucket := range dynamicBuckets {
		dynamicBucket.RemoveListeners(bucketListener)
	}
}

func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	prefix = strings.TrimPrefix(prefix, "/")
	prefixes := strings.Split(prefix, ",")
	dynamicBuckets, err := c.getDynamicBuckets(prefixes)
	if err != nil {
		return waitIndex, err
	}

	if waitIndex == 0 {
		return waitIndex+1, nil
	}  else {
		watchResp := make(chan *watchResponse)
		bucketListener := &BucketListener{watchResp: watchResp, currentIndex: waitIndex}
		setupDynamicBucketListeners(dynamicBuckets, bucketListener)
		select {
			case watchResp := <- watchResp:
				removeDynamicBucketListeners(dynamicBuckets, bucketListener)
		 		return watchResp.waitIndex, watchResp.err
		    case <-stopChan:
				removeDynamicBucketListeners(dynamicBuckets, bucketListener)
				return 0, nil
		}
	}
}

