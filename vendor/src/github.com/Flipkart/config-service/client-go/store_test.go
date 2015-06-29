package cfgsvc
import (
    "net/http"
    "testing"
    "github.com/stretchr/testify/assert"
)

func Test_HttpClient_GetBucket(t *testing.T) {

    server, httpClient := httpTestTool(200, `{"metadata":{"name":"foo","version":10,"lastUpdated":1431597657},"keys":{"foo":"bar","bar":"baz"}}`)
    client, err := NewConfigServiceClient(server.URL, 50)
    client.httpClient, _ = NewHttpClient(&httpClient, server.URL)
    assert.Nil(t,err)
    assert.NotNil(t,client)

    b, err := client.GetBucket("foo", -1)
    assert.Nil(t,err)
    assert.NotNil(t, b)

    testBucket(t, b)

}

func Test_HttpClient_WatchBucket(t *testing.T) {

}

func Test_HttpClient_Construction(t *testing.T) {
    _, err := NewHttpClient(&http.Client{}, "http://localhost:8080")
    assert.Nil(t, err)
}

