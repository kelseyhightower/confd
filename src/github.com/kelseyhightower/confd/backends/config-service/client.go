package config_service

import (
	"strings"
	cfgsvc "github.com/Flipkart/config-service/client-go"
	"github.com/kelseyhightower/confd/log"
	"errors"
)

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *cfgsvc.ConfigServiceClient
	bucketListener *BucketListener
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
	this.watchResp <- &watchResponse{waitIndex:uint64(newBucket.GetVersion()+1), err: nil}
}


func NewConfigClient(machines []string) (*Client, error) {
	c, err := cfgsvc.NewConfigServiceClient(machines[0], 50) //*10)
	if err != nil {
		panic(err)
	}
	return &Client{client:c, bucketListener: &BucketListener{watchResp: make(chan *watchResponse), currentIndex: 1}}, nil
}


func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range keys {
		bucketKeys := strings.Split(v[1:], "/")
		bucket, err := c.client.GetDynamicBucket(bucketKeys[0])
		if err != nil {
			return vars, err
		}

		val := bucket.GetKeys()[bucketKeys[1]]
		if val != nil {
			value := val.(string)
			vars[v] = value
		}

	}
	return vars, nil
}


func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	dynamicBucket, err := c.client.GetDynamicBucket(strings.TrimPrefix(prefix, "/"))
	if err != nil {
		return waitIndex, err
	}

	if waitIndex == 0 {
		dynamicBucket.AddListeners(c.bucketListener)
		return uint64(dynamicBucket.GetVersion() +1), nil
	}  else {
		select {
			case watchResp := <- c.bucketListener.watchResp:
		 		return watchResp.waitIndex, watchResp.err
		    case <-stopChan:
				return 0, nil
		}
	}
}

