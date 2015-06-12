package cfgsvc
import (
	"errors"
	"github.com/pquerna/ffjson/ffjson"
    "strconv"
    "regexp"
)

/**
 * Bucket of config keys
 **/

//Bucket instance which exposes API for property reads and watches.
type bucket struct {
	Meta *BucketMetaData           `json:"metadata"`
	Keys map[string]interface{}    `json:"keys"`
}
type Bucket struct {
	bucket
}

/** Getters */

func (this *Bucket) GetMeta() *BucketMetaData {
	return this.Meta
}
func (this *Bucket) GetKeys() map[string]interface{} {
    return this.Keys
}

/** BucketInterface implementations */

func (this *Bucket) GetName() string {
    return this.Meta.GetName()
}
func (this *Bucket) GetVersion() uint {
	meta := this.Meta
	if meta != nil {
		return meta.GetVersion()
	} else {
		return 0
	}
}
func (this *Bucket) GetLastUpdated() uint64 {
    return this.Meta.GetLastUpdated()
}


/** Type specific getters */

func (this *Bucket) GetBool(name string) (bool, error) {
    if val, ok := (this.Keys[name]).(bool); ok {
        return val, nil
    } else {
        return false, errors.New("Not a boolean value")
    }
}
func (this *Bucket) GetString(name string) (string, error){
    if val, ok := (this.Keys[name]).(string); ok {
        return val, nil
    } else {
        return "", errors.New("Not a string value")
    }
}
func (this *Bucket) GetInt(name string) (int, error) {
    if val, ok := (this.Keys[name]).(int); ok {
        return val, nil
    } else {
        return 0, errors.New("Not a integer value")
    }
}
func (this *Bucket) GetFloat(name string) (float64, error) {
    if val, ok := (this.Keys[name]).(float64); ok {
        return val, nil
    } else {
        return 0.0, errors.New("Not a float value")
    }
}
func (this *Bucket) GetBoolArray(name string) ([]bool, error) {
	if val, ok := (this.Keys[name]).([]bool); ok {
		return val, nil
	} else {
		return []bool {false}, errors.New("Not a boolean array")
	}
}
func (this *Bucket) GetStringArray(name string) ([]string, error){
	if val, ok := (this.Keys[name]).([]string); ok {
		return val, nil
	} else {
		return []string {"avc"}, errors.New("Not a string array")
	}
}
func (this *Bucket) GetIntArray(name string) ([]int, error) {
	if val, ok := (this.Keys[name]).([]int); ok {
		return val, nil
	} else {
		return []int{0}, errors.New("Not a integer array")
	}
}
func (this *Bucket) GetFloatArray(name string) ([]float64, error) {
	if val, ok := (this.Keys[name]).([]float64); ok {
		return val, nil
	} else {
		return []float64{0.0}, errors.New("Not a float array")
	}
}

/** JSON conversion */

func (this *Bucket) MarshalJSON() ([]byte, error) {
    return ffjson.Marshal(this.bucket)
}
func (this *Bucket) UnmarshalJSON(b []byte) error {
    return ffjson.Unmarshal(b, &this.bucket)
}

/** stinger */

func (this *Bucket) String() string {
    str,err := ffjson.Marshal(this)
    if err != nil {
        return "Error encoding to JSON: " + err.Error()
    } else {
        return string(str)
    }
}

/** methods used in sidekick for etag generation */

func (this *Bucket) GetVersionAsString() string {
    version := this.GetVersion()
    versionStr := strconv.FormatUint(uint64(version), 10)
    return versionStr
}

//Unique identification of dynamic bucket.
func (this *Bucket) GetId() string {
    return this.GetMeta().GetName() + this.GetVersionAsString() + strconv.FormatUint(this.GetMeta().GetLastUpdated(), 10)
}

/** private methods required for client-go */

func (this *Bucket) isValid() bool {
    return this.GetMeta() != nil
}

func ValidateBucketName(bucketName string) error {
    r, err := regexp.Compile(`^[a-zA-Z0-9._-]+$`)
    if err != nil {
        return err;
    }
    if bucketName == "" ||
    !r.MatchString(bucketName) {
        return errors.New("Bucket name is invalid")
    }
    return nil
}
