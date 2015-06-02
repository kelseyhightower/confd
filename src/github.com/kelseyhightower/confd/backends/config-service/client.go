package config_service

import (
	"strings"
	cfgsvc "github.com/Flipkart/config-service/client-go"
	"github.com/kelseyhightower/confd/log"
	"errors"
	"fmt"
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
		bucketKeys := strings.Split(v[1:], "/")

		dynamicBuckets, err := c.getDynamicBuckets(strings.Split(bucketKeys[0], ","))
		if err != nil {
			return vars, err
		}


		for _, dynamicBucket := range dynamicBuckets {
			val := dynamicBucket.GetKeys()[bucketKeys[1]]
			if val != nil {
				vars[bucketKeys[1]] = fmt.Sprint(val)
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

