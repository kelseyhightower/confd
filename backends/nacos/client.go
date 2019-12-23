package nacos

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/confd/log"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var replacer = strings.NewReplacer("/", ".")

type Client struct {
	nodes   []string
	group   string
	channel chan int
}
type ConfigValue struct {
	key     string
	md5     string
	content string
}

var configCache = map[string]*ConfigValue{}

func NewNacosClient(nodes []string, group string) (client *Client, err error) {
	client = &Client{nodes, group, make(chan int, 10)}
	return
}
func (client *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		k := strings.TrimPrefix(key, "/")
		k = replacer.Replace(k)
		// get instance list api when key has prefix naming
		if strings.HasPrefix(k, "naming:") {
			k = strings.TrimPrefix(k, "naming:")
			// if cache exists
			oldConfig, exist := configCache[k]
			if exist {
				vars[key] = oldConfig.content
				continue
			}
			path := fmt.Sprintf("/nacos/v1/ns/instance/list?healthyOnly=false&groupName=%s&serviceName=%s", client.group, k)
			httpClient := &http.Client{Timeout: 5 * time.Second}
			resp, err := httpClient.Get(client.buildUrl(path))
			if err != nil {
				return vars, err
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			var instanceList InstanceList
			err = json.Unmarshal(body, &instanceList)
			if err != nil {
				return vars, err
			}
			bytes, err := json.Marshal(instanceList.Hosts)
			if err != nil {
				return vars, err
			}
			md5Val := fmt.Sprintf("%x", md5.Sum(bytes))
			newConfig := &ConfigValue{k, md5Val, string(bytes)}
			configCache[k] = newConfig
			vars[key] = newConfig.content
			log.Info("service instances updated for %s ", k)
		} else {
			config, exist := configCache[k]
			if exist {
				vars[key] = config.content
				continue
			}
			// get config
			path := fmt.Sprintf("/nacos/v1/cs/configs?group=%s&dataId=%s", client.group, k)
			httpClient := &http.Client{Timeout: 5 * time.Second}
			resp, err := httpClient.Get(client.buildUrl(path))
			if err != nil {
				return vars, err
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			md5Val := resp.Header.Get("Content-MD5")
			newVal := &ConfigValue{k, md5Val, string(body)}
			configCache[k] = newVal
			log.Info("config value updated for %s %", k, newVal.content)
			vars[key] = newVal.content
		}
	}
	return vars, nil
}
func (client *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		return 1, nil
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	for {
		for _, key := range keys {
			k := strings.TrimPrefix(key, "/")
			k = replacer.Replace(k)
			if strings.HasPrefix(k, "naming:") {
				k = strings.TrimPrefix(k, "naming:")
				path := fmt.Sprintf("/nacos/v1/ns/instance/list?healthyOnly=false&groupName=%s&serviceName=%s", client.group, k)
				httpClient := &http.Client{Timeout: 5 * time.Second}
				resp, err := httpClient.Get(client.buildUrl(path))
				if err != nil {
					continue
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				var instanceList InstanceList
				err = json.Unmarshal(body, &instanceList)
				if err != nil {
					continue
				}
				bytes, err := json.Marshal(instanceList.Hosts)
				if err != nil {
					continue
				}
				md5Val := fmt.Sprintf("%x", md5.Sum(bytes))
				oldConfig, exist := configCache[k]
				if exist && strings.Compare(md5Val, oldConfig.md5) != 0 {
					log.Info("instances of service [%s] has changed", k)
					newConfig := &ConfigValue{k, md5Val, string(bytes)}
					configCache[k] = newConfig
					return waitIndex, nil
				} else {
					newConfig := &ConfigValue{k, md5Val, string(bytes)}
					configCache[k] = newConfig
				}
			} else {
				body := "Listening-Configs="
				body += k + string(2)
				body += client.group + string(2)
				body += configCache[k].md5 + string(2)
				body += "" + string(1)
				// long pulling for listener
				path := "/nacos/v1/cs/configs/listener"
				req, _ := http.NewRequest("POST", client.buildUrl(path), strings.NewReader(body))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Long-Pulling-Timeout", "100")
				resp, err := httpClient.Do(req)
				if err != nil {
					continue
				}
				defer resp.Body.Close()
				respBody, _ := ioutil.ReadAll(resp.Body)
				// get result and it's not empty
				result := string(respBody)
				if len(result) > 0 {
					log.Info("config [%s] has changed", k)
					delete(configCache, k)
					return waitIndex, nil
				}
			}
		}
		time.Sleep(time.Microsecond * 500)
	}
	return waitIndex, nil
}

func (client *Client) buildUrl(path string) string {
	n := rand.Intn(len(client.nodes))
	node := client.nodes[n]
	return node + path
}

type InstanceList struct {
	Name  string         `json:"name"`
	Hosts []InstanceHost `json:"hosts"`
}
type InstanceHost struct {
	InstanceId  string            `json:"instanceId"`
	ServiceName string            `json:"serviceName"`
	Ip          string            `json:"ip"`
	Port        int               `json:"port"`
	Weight      float32           `json:"weight"`
	Enabled     bool              `json:"enabled"`
	Healthy     bool              `json:"healthy"`
	Valid       bool              `json:"valid"`
	Metadata    map[string]string `json:"metadata"`
}
