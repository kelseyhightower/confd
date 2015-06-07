package cfgsvc
import (
    "github.com/pquerna/ffjson/ffjson"
)


/** Bucket's metadata */
type bucketMetaData struct {
    Name string         `json:"name"`
    Version uint        `json:"version"`
    LastUpdated uint64  `json:"lastUpdated"`
}
type BucketMetaData struct {
    bucketMetaData
}
func (this *BucketMetaData) GetName() string {
    return this.Name
}
func (this *BucketMetaData) GetVersion() uint {
    return this.Version
}
func (this *BucketMetaData) GetLastUpdated() uint64 {
    return this.LastUpdated
}


/** JSON conversion */
func (this *BucketMetaData) MarshalJSON() ([]byte, error) {
    return ffjson.Marshal(this.bucketMetaData)
}
func (this *BucketMetaData) UnmarshalJSON(b []byte) error {
    return ffjson.Unmarshal(b, &this.bucketMetaData)
}


/** stringer */
func (this *BucketMetaData) String() string {
    str,err := ffjson.Marshal(this)
    if err != nil {
        return "Error encoding to JSON: " + err.Error()
    } else {
        return string(str)
    }
}



type BucketInterface interface {

    /** getters */

    GetMeta() *BucketMetaData
    GetKeys() map[string]interface{}
    GetName() string
    GetVersion() uint
    GetLastUpdated() uint64

    /** used for etag calculation */

    GetId() string

    /** type specific getters */

    GetBool(string) (bool, error)
    GetString(name string) (string, error)
    GetInt(name string) (int, error)
    GetFloat(name string) (float64, error)
    GetBoolArray(name string) ([]bool, error)
    GetStringArray(name string) ([]string, error)
    GetIntArray(name string) ([]int, error)
    GetFloatArray(name string) ([]float64, error)

}

/*
  BucketUpdatesListener is an interface to be implemented for
  capturing events related to bucket data changes and errors
 */
type BucketUpdatesListener interface {
    //Callback made when a bucket is updated. Old and new
    //versions of the bucket are provided for comparison purposes.

    //The provided instances of buckets are static in nature, ie,
    //they are not auto-updated by next watch request.
    Updated(oldBucket *Bucket, newBucket *Bucket)

    //Callback made when bucket is deleted.
    Deleted(bucketName string)

    //Callback made when watch is disconnected
    //the service.
    Disconnected(bucketName string, err error)

    //Callback made when watch is connected
    //the service.
    Connected(bucketName string)
}

