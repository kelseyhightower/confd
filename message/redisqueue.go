package message

// use redis cmd lpush to message
import (
	"os"
	"bufio"
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/mafengwo/confd/log"
)

//init redis client
func newRedisClient(redisconf string) redis.Conn {
	RedisModel, err := redis.Dial("tcp", redisconf)
	RedisModel.Do("select", "0")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("connect redis " + redisconf + ":0")
	return RedisModel
}

//send message
func SendMessage(redisconf, dest string) {
	RedisModel := newRedisClient(redisconf)
	defer RedisModel.Close()

	key := "server_config_queue"
	body := body(dest)
	if body == "" {
		log.Info("redis value is empty!")
	}
	//send message to redis
	_, err := RedisModel.Do("lpush", key, body)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("value is " + body)
}

//json message
func body(dest string) string {
	content := make([]string, 0)
	output := make(map[string]interface{})

	file, err := os.Open(dest)
    if err!= nil {
     	log.Fatal(err.Error())
        return ""
    }
    defer file.Close()
    reader := bufio.NewReader(file)
    for {
        str, err := reader.ReadString('\n') 
        if err != nil {
            break 
        }
        content = append(content, str)
    }
    output["content"] = content
    output["dest"] = dest
    output["timestamp"] = time.Now().Unix()
    output_json, err := json.Marshal(output)
    if err != nil {
    	log.Fatal(err.Error())
    	return ""
    }
    return string(output_json)
}

