package message

// use redis cmd lpush to message
import (
	"os"
	"bufio"
	"encoding/json"
	"time"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/mafengwo/confd/log"
)

//init redis client
func newRedisClient(redisconf string) (redis.Conn, error) {
	RedisModel, err := redis.Dial("tcp", redisconf)
	return RedisModel, err
}

//send message
func SendMessage(redisconf, dest string) bool {
	RedisModel, err := newRedisClient(redisconf)
	if err != nil {
		var i int
		log.Info("connected redis " + redisconf + " failed!")
		time.Sleep(1*time.Second)
		//retry connect redis
		for i = 1; i <= 3; i++ {
			RedisModel, err = newRedisClient(redisconf)
			if err != nil {
				log.Info("reconnected redis " + redisconf + " failed! " + strconv.Itoa(i) + " times")	
				time.Sleep(1*time.Second)	
			} else {
				break
			}
		}
		if i == 4 {
			log.Info("reconnected redis " + redisconf + " all failed!")
			return false
		}
	} 
	log.Info("connected redis " + redisconf + " successful!")
	defer RedisModel.Close()
	key := "server_config_queue"
	body := body(dest)
	if body == "" {
		log.Info("redis value is empty!")
	}
	//send message to redis
	_, err = RedisModel.Do("lpush", key, body)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("value is " + body)
	return true
}

//json message
func body(dest string) string {
	content := make([]string, 0)
	output := make(map[string]interface{})

	file, err := os.Open(dest)
    if err!= nil {
    	log.Info("open dest file failed!")
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
    	log.Info("json_encode failed!")
    }
    return string(output_json)
}

