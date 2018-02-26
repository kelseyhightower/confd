package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	"sync"
)

// A watch only tells the latest revision
type Watch struct {
	// Last seen revision
	revision int64
	// A channel to wait, will be closed after revision changes
	cond chan bool
	// Use RWMutex to protect cond variable
	rwl sync.RWMutex
}

// Wait until revision is greater than lastRevision
func (w *Watch) WaitNext(ctx context.Context, lastRevision int64, notify chan<-int64) {
	for {
		w.rwl.RLock()
		if w.revision > lastRevision {
			w.rwl.RUnlock()
			break
		}
		cond := w.cond
		w.rwl.RUnlock()
		select{
		case <-cond:
		case <-ctx.Done():
			return
		}
	}
	// We accept larger revision, so do not need to use RLock
	select{
	case notify<-w.revision:
	case <-ctx.Done():
	}
}

// Update revision
func (w *Watch) update(newRevision int64){
	w.rwl.Lock()
	defer w.rwl.Unlock()
	w.revision = newRevision
	close(w.cond)
	w.cond = make(chan bool)
}

func createWatch(client *clientv3.Client, prefix string) (*Watch, error) {
	w := &Watch{0, make(chan bool), sync.RWMutex{}}
	go func() {
		rch := client.Watch(context.Background(), prefix, clientv3.WithPrefix())
		for {
			for wresp := range rch {
				if wresp.CompactRevision > w.revision {
					// respect CompactRevision
					w.update(wresp.CompactRevision)
				} else if wresp.Header.GetRevision() > w.revision {
					// Watch created or updated
					w.update(wresp.Header.GetRevision())
				}
			}
			// Disconnected or cancelled
			// Wait for a moment to avoid reconnecting
			// too quickly
			time.Sleep(time.Duration(1) * time.Second)
			// Start from next revision so we are not missing anything
			if w.revision > 0 {
				rch = client.Watch(context.Background(), prefix, clientv3.WithPrefix(),
					clientv3.WithRev(w.revision + 1))
			} else {
				// Start from the latest revision
				rch = client.Watch(context.Background(), prefix, clientv3.WithPrefix())
			}
		}
	}()
	return w, nil
}

// Client is a wrapper around the etcd client
type Client struct {
	client *clientv3.Client
	watches map[string]*Watch
	// Protect watch
	wm sync.Mutex
}

// NewEtcdClient returns an *etcdv3.Client with a connection to named machines.
func NewEtcdClient(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	cfg := clientv3.Config{
		Endpoints:            machines,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if caCert != "" {
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return &Client{}, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return &Client{}, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		cfg.TLS = tlsConfig
	}
	
	client, err := clientv3.New(cfg)
	if err != nil {
		return &Client{}, err
	}
	
	return &Client{client, make(map[string]*Watch), sync.Mutex{}}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	getOps := make([]clientv3.Op, 0)
	for _, key := range keys {
		getOps = append(getOps, clientv3.OpGet(key,
											   clientv3.WithPrefix(),
											   clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend)))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3) * time.Second)
	result, err := c.client.Txn(ctx).Then(getOps...).Commit()
	cancel()
	if err != nil {
		return vars, err
	}
	for _, r := range result.Responses {
		for _, ev := range r.GetResponseRange().Kvs {
			vars[string(ev.Key)] = string(ev.Value)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	var err error
	
	// Create watch for each key
	watches := make(map[string]*Watch)
	c.wm.Lock()
	for _, k := range keys {
		watch, ok := c.watches[k]
		if !ok {
			watch, err = createWatch(c.client, k)
			if err != nil {
				c.wm.Unlock()
				return 0, err
			}
			c.watches[k] = watch
		}
		watches[k] = watch
	}
	c.wm.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancelRoutine := make(chan bool)
	defer cancel()

	go func() {
		select {
		case <-stopChan:
			cancel()
		case <-cancelRoutine:
			return
		}
	}()
	
	notify := make(chan int64)
	// Wait for all watches
	for _, v := range watches {
		go v.WaitNext(ctx, int64(waitIndex), notify)
	}
	select{
	case nextRevision := <- notify:
		return uint64(nextRevision), err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	return 0, err
}
