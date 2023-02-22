package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"net/url"
	"strconv"
	"strings"
)

var replacer = strings.NewReplacer("/", ".")

type Client struct {
	client    config_client.IConfigClient
	group     string
	namespace string
	accessKey string
	secretKey string
	channel   chan int
}

func NewNacosClient(nodes []string, group string, config constant.ClientConfig) (client *Client, err error) {
	var configClient config_client.IConfigClient
	servers := []constant.ServerConfig{}
	for _, key := range nodes {
		nacosUrl, _ := url.Parse(key)

		fmt.Println(key)
		fmt.Println(nacosUrl.Hostname())
		fmt.Println(nacosUrl.Port())
		port, _ := strconv.Atoi(nacosUrl.Port())
		servers = append(servers, constant.ServerConfig{
			IpAddr: nacosUrl.Hostname(),
			Port:   uint64(port),
		})
	}

	fmt.Println("namespace=" + config.NamespaceId)
	fmt.Println("AccessKey=" + config.AccessKey)
	fmt.Println("SecretKey=" + config.SecretKey)
	fmt.Println("Endpoint=" + config.Endpoint)

	configClient, err = clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": servers,
		"clientConfig": constant.ClientConfig{
			TimeoutMs:           20000,
			ListenInterval:      10000,
			NotLoadCacheAtStart: true,
			NamespaceId:         config.NamespaceId,
			AccessKey:           config.AccessKey,
			SecretKey:           config.SecretKey,
			Endpoint:            config.Endpoint,
		},
	})

	if len(strings.TrimSpace(group)) == 0 {
		group = "DEFAULT_GROUP"
	}

	namespace := strings.TrimSpace(config.NamespaceId)

	client = &Client{configClient, group, namespace, config.AccessKey, config.SecretKey, make(chan int)}
	fmt.Println("hello nacos")

	return
}

func (client *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	fmt.Println(keys)
	for _, key := range keys {
		k := strings.TrimPrefix(key, "/")
		k = replacer.Replace(k)
		resp, err := client.client.GetConfig(vo.ConfigParam{
			DataId: k,
			Group:  client.group,
		})
		fmt.Println(k + ":" + resp)
		if err == nil {
			vars[key] = resp
		}
	}

	return vars, nil
}

func (client *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		for _, key := range keys {
			k := strings.TrimPrefix(key, "/")
			k = replacer.Replace(k)

			err := client.client.ListenConfig(vo.ConfigParam{
				DataId: k,
				Group:  client.group,
				OnChange: func(namespace, group, dataId, data string) {
					fmt.Println(data)
					client.channel <- 1
				},
			})
			fmt.Println(key)

			if err != nil {
				return 0, err
			}
		}

		return 1, nil
	}

	select {
	case <-client.channel:
		return waitIndex, nil

	}
	fmt.Print("waitIndex=")
	fmt.Println(waitIndex)
	fmt.Println(prefix)
	fmt.Println(keys)

	return waitIndex, nil
}
