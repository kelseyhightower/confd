package zookeeper

import (
        "time"
        "strings"

	zk "github.com/samuel/go-zookeeper/zk"
)



// Client provides a wrapper around the zookeeper client
type Client struct{
    client *zk.Conn;
}


func NewZookeeperClient(machines []string) (*Client, error) {
        c, _, err := zk.Connect(machines, time.Second) //*10)
        if err != nil {
		panic(err)
	}
        return &Client{c}, nil
}


func node_walk(prefix string, c *Client, vars map[string]string ) error{
    l,stat,err:= c.client.Children(prefix);
        if err != nil {
                 return err
        }

    if (stat.NumChildren == 0) {
        b,_, err:= c.client.Get(prefix);
        if err != nil {
                 return err
        }
        vars[prefix]=string(b)
      
    } else {
    for _, key := range l {
       s := prefix+"/"+key
       _, stat, err := c.client.Exists(s);
        if err != nil {
                 return err
        }
        if (stat.NumChildren == 0) {
             b,_, err:= c.client.Get(s);
        if err != nil {
                 return err
        }
             vars[s]=string(b)
        } else {
          node_walk(s,c,vars)
        }
    }
   }
   return nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
    vars := make(map[string]string)
    for _, v := range keys {
        v= strings.Replace(v,"/*","",-1);
        _, _, err := c.client.Exists(v);
        if err != nil {
	         return vars, err
	}
          if (v == "/" ) {v=""}
          err=node_walk(v,c,vars)
		if err != nil {
			return vars, err
		}
     }
    return vars, nil
}


// WatchPrefix is not yet implemented. There's a WIP.
// Since zookeeper doesn't handle recursive watch, we need to create a *lot* of watches.
// Implementation should take care of this.
// A good start is bamboo
// URL https://github.com/QubitProducts/bamboo/blob/master/qzk/qzk.go
// We also need to encourage users to set prefix and add a flag to enale support for "" prefix (aka "/")
//

func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}

