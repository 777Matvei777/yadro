package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type XkcdStruct struct {
	Id         int    `json:"num"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Url        string `json:"img"`
}

func Parse(Url string, Parallel int, ctx context.Context, num int, exist map[int]bool) []XkcdStruct {
	var Db []XkcdStruct
	var wg sync.WaitGroup
	var mutex sync.Mutex
	ch := make(chan int, Parallel)
	flag := false
	for i := num; !flag; i++ {
		select {
		case <-ctx.Done():
			return Db
		default:
			if !exist[i] {
				ch <- i
				wg.Add(1)
				i := i
				go func(i int) {
					defer wg.Done()
					defer func() { <-ch }()

					address := fmt.Sprintf("%s/%d/info.0.json", Url, i)
					resp, err := http.Get(address)
					if err != nil {
						log.Println("Error GET comics", err)
						return
					}
					defer resp.Body.Close()
					if resp.StatusCode == 404 && i == 404 {
						return
					}
					if resp.StatusCode == 404 && i != 404 {
						mutex.Lock()
						flag = true
						mutex.Unlock()
						return
					}
					var oneData XkcdStruct
					data, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Println("Error reading responce", err)
						return
					}
					_ = json.Unmarshal([]byte(data), &oneData)
					mutex.Lock()
					Db = append(Db, oneData)
					mutex.Unlock()

				}(i)
			}

		}
	}
	wg.Wait()
	close(ch)
	return Db
}
