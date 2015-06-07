package cfgsvc
import (
	"strconv"
	"sync/atomic"
	"github.com/pquerna/ffjson/ffjson"
	"errors"
	"sync"
	"log"
	"time"
)

//Dynamic bucket is a proxy to the immutable Bucket object.
//It's used in case a bucket needs to be auto-updated.
type DynamicBucket struct {
	bucket *Bucket
	stopWatch chan bool
	lastChecked int64
	lock sync.RWMutex
	listeners []BucketUpdatesListener
	httpClient *HttpClient
}

/** lifecycle methods */

//Initialize the dynamic bucket by fetching the current bucket
//and setting up watch for further updates.
func (this *DynamicBucket) init(name string) error {
	this.stopWatch = make(chan bool, 1)
	this.lastChecked = -1
	this.lock = sync.RWMutex{}
	bucket, err := this.httpClient.GetBucket(name, LATEST_VERSION)

	if err != nil {
		log.Println("Error fetching bucket: ", err)
		return err
	}
	this.updateBucketWithoutCallbacks(bucket)
	return nil
}

//Shutdown the dynamic bucket by stopping the watch.
func (this *DynamicBucket) shutdown() {
    close(this.stopWatch)
}

//Checks if dynamic bucket watch is running or not.
func (this *DynamicBucket) isShutdown() (chan bool) {
    return this.stopWatch
}

/** bucket operations */
//Get the immutable bucket.
func (this *DynamicBucket) getBucket() (*Bucket) {
    this.lock.RLock(); defer this.lock.RUnlock()
    return this.bucket
}
//Update the static bucket.

//Also, calls update callback on all listeners.
func (this *DynamicBucket) updateBucket(newBucket *Bucket) {
	this.lock.Lock(); defer this.lock.Unlock()
	for _, listener := range this.listeners {
		listener.Updated(this.bucket, newBucket)
	}
	this.bucket = newBucket
}

//Update the static bucket without callbacks.
func (this *DynamicBucket) updateBucketWithoutCallbacks(newBucket *Bucket) {
	this.lock.Lock(); defer this.lock.Unlock()
	this.bucket = newBucket
}

/** listener operations */

//Add a new listener.
func (this *DynamicBucket) AddListeners(listener BucketUpdatesListener) {
	this.lock.Lock(); defer this.lock.Unlock()
	this.listeners = append(this.listeners, listener)
}

//Remove a listener.
func (this *DynamicBucket) RemoveListeners(listener BucketUpdatesListener) {
	this.lock.Lock(); defer this.lock.Unlock()
	for i, item := range this.listeners {
		if item == listener {
			this.listeners = append(this.listeners[:i], this.listeners[i+1:]...)
			break
		}
	}
}

//Delete the static bucket.
func (this *DynamicBucket) DeleteBucket() {
	this.lock.RLock(); defer this.lock.RUnlock()
	for _, listener := range this.listeners {
		listener.Deleted(this.GetMeta().GetName())
	}
}

//Disconnected callback.
func (this *DynamicBucket) Disconnected(err error) {
	this.lock.RLock(); defer this.lock.RUnlock()
	if (this.isConnected()) {
		this.SetLastChecked(time.Now().Unix())
		for _, listener := range this.listeners {
			listener.Disconnected(this.GetMeta().GetName(), err)
		}
	}
}

//Connected callback.
func (this *DynamicBucket) Connected() {
	this.lock.RLock(); defer this.lock.RUnlock()
	if (!this.isConnected()) {
		this.SetLastChecked(-1)
		for _, listener := range this.listeners {
			listener.Connected(this.GetMeta().GetName())
		}
	}

}

//Checks if watch is connected or not
func (this *DynamicBucket) isConnected() bool {
	if (this.GetLastChecked() == -1)  {
		return true
	}
	return false
}

/** Utilities */

func (this *DynamicBucket) GetLastChecked() (int64) {
    return atomic.LoadInt64(&this.lastChecked)
}
func (this *DynamicBucket) SetLastChecked(timestamp int64) {
    atomic.StoreInt64(&this.lastChecked, timestamp)
}

/** methods used in sidekick for etag generation */

func (this *DynamicBucket) GetVersionAsString() string {
    version := this.GetVersion()
    versionStr := strconv.FormatUint(uint64(version), 10)
    return versionStr
}

//Unique identification of dynamic bucket.
func (this *DynamicBucket) GetId() string {
    return this.GetMeta().GetName() + this.GetVersionAsString() + strconv.FormatUint(this.GetMeta().GetLastUpdated(), 10)
}

/** BucketInterface implementations */

func (this *DynamicBucket) GetMeta() *BucketMetaData {
	return this.getBucket().Meta
}
func (this *DynamicBucket) GetKeys() map[string]interface{} {
    return this.getBucket().Keys
}
func (this *DynamicBucket) GetName() string {
    return this.GetMeta().GetName()
}
func (this *DynamicBucket) GetVersion() uint {
	meta := this.GetMeta()
	if meta != nil {
		return meta.GetVersion()
	} else {
		return 0
	}
}
func (this *DynamicBucket) GetLastUpdated() uint64 {
    return this.GetMeta().GetLastUpdated()
}


/** Type specific getters */

func (this *DynamicBucket) GetBool(name string) (bool, error) {
	if val, ok := (this.GetKeys()[name]).(bool); ok {
		return val, nil
	} else {
		return false, errors.New("Not a boolean value")
	}
}
func (this *DynamicBucket) GetString(name string) (string, error){
	if val, ok := (this.GetKeys()[name]).(string); ok {
		return val, nil
	} else {
		return "", errors.New("Not a string value")
	}
}
func (this *DynamicBucket) GetInt(name string) (int, error) {
	if val, ok := (this.GetKeys()[name]).(int); ok {
		return val, nil
	} else {
		return 0, errors.New("Not a integer value")
	}
}
func (this *DynamicBucket) GetFloat(name string) (float64, error) {
	if val, ok := (this.GetKeys()[name]).(float64); ok {
		return val, nil
	} else {
		return 0.0, errors.New("Not a float value")
	}
}

func (this *DynamicBucket) GetBoolArray(name string) ([]bool, error) {
	if val, ok := (this.GetKeys()[name]).([]bool); ok {
		return val, nil
	} else {
		return []bool{false}, errors.New("Not a boolean array")
	}
}
func (this *DynamicBucket) GetStringArray(name string) ([]string, error){
	if val, ok := (this.GetKeys()[name]).([]string); ok {
		return val, nil
	} else {
		return []string{""}, errors.New("Not a string array")
	}
}
func (this *DynamicBucket) GetIntArray(name string) ([]int, error) {
	if val, ok := (this.GetKeys()[name]).([]int); ok {
		return val, nil
	} else {
		return []int{0}, errors.New("Not a integer array")
	}
}
func (this *DynamicBucket) GetFloatArray(name string) ([]float64, error) {
	if val, ok := (this.GetKeys()[name]).([]float64); ok {
		return val, nil
	} else {
		return []float64{0.0}, errors.New("Not a float array")
	}
}

/** JSON conversion */

func (this *DynamicBucket) MarshalJSON() ([]byte, error) {
	return ffjson.Marshal(this.getBucket())
}
func (this *DynamicBucket) UnmarshalJSON(b []byte) error {
	return ffjson.Unmarshal(b, this.getBucket())
}

/** stringer */

func (this *DynamicBucket) String() string {
	str,err := ffjson.Marshal(this)
	if err != nil {
		return "Error encoding to JSON: " + err.Error()
	} else {
		return string(str)
	}
}

