package cfgsvc
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "net/http"
    "net/url"
    "fmt"
    "net/http/httptest"
)


var (
    meta BucketMetaData = BucketMetaData{
        bucketMetaData: bucketMetaData{
            Name: "foo",
            Version: 10,
            LastUpdated: 1431597657,
        },
    }
    testBucketData Bucket = Bucket{
       bucket: bucket {
           Meta: &meta,
           Keys: map[string]interface{}{
                "foo":"bar",
                "bar": "baz",
           },
       },
    }
)

func testBucket(t *testing.T, b *Bucket) {
    testMeta(t, b.GetMeta())
    testProperty(t, b.GetKeys())
}

func testMeta(t *testing.T, m *BucketMetaData) {
    assert.Equal(t, m.GetName(), meta.GetName())
    assert.Equal(t, m.GetVersion(), meta.GetVersion())
    assert.Equal(t, m.GetLastUpdated(), meta.GetLastUpdated())
}

func testProperty(t *testing.T, p map[string]interface{}) {
    assert.Equal(t, p, testBucketData.GetKeys())
}


func Test_ConmanClient_GetBucket(t *testing.T) {

    server, httpClient := httpTestTool(200, `{"metadata":{"name":"foo","version":10,"lastUpdated":1431597657},"keys":{"foo":"bar","bar":"baz"}}`)
    client, err := NewConfigServiceClient(server.URL, 50)
    client.httpClient, _ = NewHttpClient(&httpClient, server.URL)
    assert.Nil(t, err)
    assert.NotNil(t, client)

    bucket, err := client.GetBucket(meta.GetName(), -1)
    assert.Nil(t, err)
    assert.NotNil(t, bucket)

    testBucket(t, bucket)

}

func httpTestTool(code int, body string) (*httptest.Server, http.Client) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(code)
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintln(w, body)
    }))
    tr := &http.Transport{
        Proxy: func(req *http.Request) (*url.URL, error) {
            return url.Parse(server.URL)
        },
    }
    client := http.Client{Transport: tr}
    return server, client
}
