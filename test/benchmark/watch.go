package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var success int32
var fail int32
var total int32
var clientNum = 10000
var wg = sync.WaitGroup{}
var client = &http.Client{}
var start time.Time

func main() {
	u := "http://127.0.0.1:30110/v1/default/kie/kv?label=app:default&label=env:test&wait=5m"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Panic(err)
		return
	}
	uri, err := url.Parse(u)
	if err != nil {
		log.Panic(err)
	}
	wg.Add(clientNum)
	client = &http.Client{
		Timeout: 1 * time.Minute,
	}
	if uri.Scheme == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	for n := 0; n < clientNum; n++ {
		go func() {
			defer wg.Done()
			err = watch(req)
			if err != nil {
				atomic.AddInt32(&fail, 1)
				return
			}
			atomic.AddInt32(&success, 1)
			if total == 0 {
				start = time.Now()
				atomic.AddInt32(&total, 1)
			}
		}()

	}

	wg.Wait()
	duration := time.Since(start)
	log.Printf("success %d", success)
	log.Printf("fail %d", fail)
	log.Printf("takes %s", duration.String())
}

func watch(req *http.Request) error {
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.Status != "200 OK" {
		log.Println(res.Status)
		return errors.New("not OK")
	}

	return nil
}
