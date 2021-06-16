package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voioc/coco/logcus"
)

// ProxyClient 类型
type ProxyClient struct {
	C *gin.Context
}

func NewProxy(c *gin.Context) ProxyClient {
	return ProxyClient{C: c}
}

//定义并初始化客户端变量
var client *http.Client

func getClinet() *http.Client {
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 信任所有证书
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, time.Second*2) // 限制建立TCP连接的时间
					if err != nil {
						return nil, err
					}
					return c, nil

				},
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second * 2, // 限制读取response header的时间
			},
			Timeout: 5 * time.Second, // 从连接(Dial)到读完response body 的时间
		}
	}

	return client
}

// HttpResponse 请求结果数据结构
type HttpResponse struct {
	URL            string
	HttpStatus     string
	HttpStatusCode int
	ContentLength  int64
	Body           []byte
}

// NewResponse 默认返回数据结构
func NewResponse() *HttpResponse {
	return &HttpResponse{HttpStatusCode: -1, Body: []byte("{}")}
}

// HttpModel 并发请求单个请求类型
type HttpModel struct {
	NeedCache  bool
	Rtype      string
	Method     string
	URL        string
	Header     map[string]string
	Postdata   map[string]string
	HTTPUniqid string
	Response   HttpResponse
}

type Result struct {
	Job  HttpModel
	Data []byte
}

func SampleClient(urls string, method string, header map[string]string, postdata interface{}) *HttpResponse {
	// Service.Flagtime("1")
	var pbody io.Reader
	req, err := http.NewRequest(method, urls, nil)
	if err != nil {
		logcus.Error("CacheHTTP gen newRequest:" + err.Error())
	}

	if postdata != nil {
		if strings.ToUpper(method) == "GET" {
			if post, ok := postdata.(map[string]string); ok {
				q := req.URL.Query()
				for k, v := range post {
					q.Add(k, v)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Common.SetDebug(fmt.Sprintf("Send HTTP Query: %s", urls+"?"+req.URL.RawQuery), 2)

		} else if strings.ToUpper(method) == "POST" {
			if post, ok := postdata.(map[string]string); ok {
				data := make(url.Values)
				for k, v := range post {
					data.Add(k, string(v))
				}
				pbody = strings.NewReader(data.Encode())
			}

			if post, ok := postdata.([]byte); ok {
				pbody = bytes.NewReader(post)
			}

			if req, err = http.NewRequest(method, urls, pbody); err != nil {
				logcus.Error("CacheHTTP gen newRequest:" + err.Error())
			}
			// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			// Common.SetDebug(fmt.Sprintf("Send HTTP Query: %s", urls), 2)
		}
	}

	//增加header
	req.Header.Add("User-Agent", "Mozilla/5.0")

	for k, v := range header {
		req.Header.Set(k, v)
	}

	httpRes := NewResponse()
	client := getClinet()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		//不抛出错误而是接口降级
		// Common.SetDebug(fmt.Sprintf("HTTP Query Downgrade: %s", err.Error()), 2)
		logcus.Error("CacheHTTP request error: " + err.Error())

		return httpRes
	}
	if resp.StatusCode != 200 {
		//不抛出错误而是接口降级
		// Common.SetDebug(fmt.Sprintf("HTTP Query Downgrade: non-200 StatusCode:%s", urls), 2)
		logcus.Error("CacheHTTP request got non-200 StatusCode: " + urls)

		httpRes.HttpStatus = resp.Status
		httpRes.HttpStatusCode = resp.StatusCode
		return httpRes
	}

	// Common.SetDebug(fmt.Sprintf("HTTP Query Result{"+Service.Flagtime("1")+"} : status :%s, content length:%d, url:%s", resp.Status, resp.ContentLength, urls), 2)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logcus.Error("CacheHTTP read response error: " + err.Error())
	}

	httpRes.URL = urls
	httpRes.HttpStatus = resp.Status
	httpRes.HttpStatusCode = resp.StatusCode
	httpRes.ContentLength = resp.ContentLength
	httpRes.Body = body

	return httpRes
}

// Multiple
type Multiple struct {
	jobs    chan HttpModel
	results chan Result
}

// NewMultiple 111
func NewMultiple() *Multiple {
	return &Multiple{jobs: make(chan HttpModel, 10), results: make(chan Result, 100)}
}

/*
* 支持多http请求并缓存，全局请求只使用同一个map channel，并将并发中的多个相同请求归并到同一个channel和responsedata
* 因为channel问题目前不暂时不支持将并发中的多个相同请求归并到同一个channel和responsedata
**/
func (m *Multiple) MultipleClient(ch []HttpModel) []Result {
	go m.allocate(ch)
	done := make(chan []Result)
	go m.result(done)
	noOfWorkers := 10
	m.createWorkerPool(noOfWorkers)
	data := <-done

	return data
}

func (m *Multiple) worker(wg *sync.WaitGroup) {
	for job := range m.jobs {
		output := Result{job, m.httpQuery(job)}
		m.results <- output
	}
	wg.Done()
}

func (m *Multiple) httpQuery(request HttpModel) []byte {
	// cache_key := "HTTP_" + request.HTTPUniqid
	// var retdata []byte
	// if request.NeedCache {
	// 	if bool, err := Cache.GetCache(cache_key, retdata); bool == true && err == nil {
	// 		return retdata
	// 	} else {
	// 		//记录log和设置debuginfo
	// 		lib.WriteLog("error", fmt.Sprintf("[error]CacheHTTP get cache:%s", err.Error()))
	// 		// Common.SetDebug(fmt.Sprintf("Cache Miss: %s", cache_key), 2)
	// 	}
	// }
	tmp := SampleClient(request.URL, request.Method, request.Header, request.Postdata)
	return tmp.Body
}

//分配协程池
func (m *Multiple) createWorkerPool(MountOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < MountOfWorkers; i++ {
		wg.Add(1)
		go m.worker(&wg)
	}
	wg.Wait()
	close(m.results)
}

/*
 * 创建任务并加入到协程池中
 */
func (m *Multiple) allocate(HttpModels []HttpModel) {
	for _, row := range HttpModels {
		m.jobs <- row
	}
	close(m.jobs)
}

/*
 * 读取返回结果
 */
func (m *Multiple) result(done chan []Result) {
	var tmp = []Result{}
	for result := range m.results {
		// if result.Job.NeedCache {
		// 	cache_key := "HTTP_" + result.Job.HTTPUniqid
		// 	if err := Cache.SetCache(cache_key, result.Data, 600); err != nil {
		// 		lib.WriteLog("error", fmt.Sprintf("[error]CacheHTTP set cache:%s", err.Error()))
		// 	}
		// 	// Common.SetDebug(fmt.Sprintf("Cache Set: %s", cache_key), 2)
		// }
		tmp = append(tmp, result)
	}
	done <- tmp
}
