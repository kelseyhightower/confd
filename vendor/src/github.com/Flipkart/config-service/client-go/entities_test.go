package cfgsvc
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func Test_Entities_BucketMeta(t *testing.T) {

    json := `{"name":"Test","version":10,"lastUpdated":1431597657}`

    meta := BucketMetaData{}
    meta.Name = "Test"
    meta.Version = 10
    meta.LastUpdated = 1431597657

    bytes,err := meta.MarshalJSON()
    assert.Nil(t,err)
    if (string(bytes) != json) {
        t.Error("JSON decode incorrect");
    }


    m := BucketMetaData{}
    m.UnmarshalJSON([]byte(json))
    if (m.GetName() != "Test" || m.GetVersion() != 10) {
        t.Error("JSON decode incorrect")
    }

}

func Test_Entities_Bucket(t *testing.T) {

    json := `{"metadata":{"name":"Test","version":10,"lastUpdated":1431597657},"keys":{"bar":"baz","foo":"bar"}}`

    meta := BucketMetaData{}
    meta.Name = "Test"
    meta.Version = 10
    meta.LastUpdated = 1431597657

    bucket := Bucket{}

    bucket.Meta = &meta
    bucket.Keys = map[string]interface{} {
        "foo":"bar",
        "bar":"baz",
    }

    bytes,err := bucket.MarshalJSON()
    assert.Nil(t,err)
    if (string(bytes) != json) {
        t.Error("Expected ", json)
        t.Error("Result ", string(bytes))
        t.Error("JSON decode incorrect ");
    }
}

func Test_Entities_DynamicBucket(t *testing.T) {

    json := `{"metadata":{"name":"Test","version":10,"lastUpdated":1431597657},"keys":{"bar":"baz","foo":"bar"}}`

    meta := BucketMetaData{}
    meta.Name = "Test"
    meta.Version = 10
    meta.LastUpdated = 1431597657

    bucket := Bucket{}

    bucket.Meta = &meta
    bucket.Keys = map[string]interface{} {
        "foo":"bar",
        "bar":"baz",
    }

    dynamicBucket := DynamicBucket{bucket:&bucket}

    bytes,err := dynamicBucket.MarshalJSON()
    assert.Nil(t,err)
    if (string(bytes) != json) {
        t.Error("Expected ", json)
        t.Error("Result ", string(bytes))
        t.Error("JSON decode incorrect ");
    }

    dynamicBucket.SetLastChecked(1431597657)
    assert.Equal(t, dynamicBucket.GetLastChecked(), int64(1431597657))

    dynamicBucket.SetLastChecked(1431597658)
    assert.Equal(t, dynamicBucket.GetLastChecked(), int64(1431597658))

    newBucket := Bucket{}

    newBucket.Meta = &meta
    newBucket.Keys = map[string]interface{} {
        "bar":"baz",
        "abce": []int {1,2, 3},
    }

    assert.NotNil(t, dynamicBucket.GetKeys()["foo"])
    dynamicBucket.updateBucket(&newBucket)
    assert.Nil(t, dynamicBucket.GetKeys()["foo"])
    val, _  := dynamicBucket.GetString("bar")
    assert.Equal(t, val, "baz")
    intVal, _ := dynamicBucket.GetIntArray("abce")
    assert.Equal(t, intVal, []int {1, 2, 3})
}

