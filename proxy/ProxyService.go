package proxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/voioc/coco/cache"
	log "github.com/voioc/coco/log"
	"github.com/voioc/coco/public"
)

// ProxyClient 类型
type ProxyClient struct {
	IsCache bool
	Debug   *[]string
}

func NewProxy() ProxyClient {
	return ProxyClient{}
}

//定义并初始化客户端变量
var client *http.Client

func getClinet() *http.Client {
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, time.Second*2)
					if err != nil {
						return nil, err
					}
					return c, nil

				},
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second * 2,
			},
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

// SampleClient 简单请求
func (p *Proxy) SampleClient(urls string, method string, header map[string]string, postdata interface{}) *HttpResponse {
	StartTime := time.Now()
	var pbody io.Reader
	req, err := http.NewRequest(method, urls, nil)
	if err != nil {
		log.Print("error", "CacheHTTP gen newRequest:", err.Error())
	}

	if postdata != nil {
		if method == "GET" || method == "get" {
			if post, ok := postdata.(map[string]string); ok {
				q := req.URL.Query()
				for k, v := range post {
					q.Add(k, v)
				}
				req.URL.RawQuery = q.Encode()
			}

			p.SetDebug(fmt.Sprintf("Send HTTP Query: %s", urls+"?"+req.URL.RawQuery), 1)

		} else if method == "POST" || method == "post" {
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
				log.Print("error", "CacheHTTP gen newRequest:", err.Error())
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			p.SetDebug(fmt.Sprintf("Send HTTP Query: %s", urls), 1)
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
		//不抛出错误而是接口降级
		p.SetDebug(fmt.Sprintf("HTTP Query Downgrade: %s", err.Error()), 2)
		log.Print("error", "CacheHTTP request:", err.Error())

		return httpRes
	}

	if resp.StatusCode != 200 {
		//不抛出错误而是接口降级
		p.SetDebug(fmt.Sprintf("HTTP Query Downgrade: non-200 StatusCode:%s", urls), 2)
		log.Print("error", "CacheHTTP request got non-200 StatusCode:", urls)

		httpRes.HttpStatus = resp.Status
		httpRes.HttpStatusCode = resp.StatusCode
		return httpRes
	}

	p.SetDebug(fmt.Sprintf("HTTP Query Result{"+public.TimeCost(StartTime)+"} : status :%s, content length:%d, url:%s", resp.Status, resp.ContentLength, urls), 1)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("panic", "CacheHTTP read response:"+err.Error())
	}

	httpRes.URL = urls
	httpRes.HttpStatus = resp.Status
	httpRes.HttpStatusCode = resp.StatusCode
	httpRes.ContentLength = resp.ContentLength
	httpRes.Body = body

	return httpRes
}

// MultipleClient 并行处理
func (p *Proxy) MultipleClient(ch []HttpModel) []Result {
	go allocate(ch)
	done := make(chan []Result)
	go p.result(done)
	noOfWorkers := 10
	p.createWorkerPool(noOfWorkers)
	data := <-done

	return data
}

var jobs = make(chan HttpModel, 10)
var results = make(chan Result, 10)

func (p *Proxy) worker(wg *sync.WaitGroup) {
	for job := range jobs {
		output := Result{job, p.httpQuery(job)}
		results <- output
	}
	wg.Done()
}

func (p *Proxy) httpQuery(request HttpModel) []byte {
	cache_key := "HTTP_" + request.HTTPUniqid
	var retdata []byte
	if request.NeedCache {
		if bool, err := cache.GetCache(cache_key, retdata); bool == true && err == nil {
			return retdata
		} else {
			//记录log和设置debuginfo
			log.Print("error", fmt.Sprintf("[error]CacheHTTP get cache:%s", err.Error()))
			p.SetDebug(fmt.Sprintf("Cache Miss: %s", cache_key), 1)
		}
	}

	tmp := p.SampleClient(request.URL, request.Method, request.Header, request.Postdata)
	return tmp.Body
}

//分配协程池
func (p *Proxy) createWorkerPool(MountOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < MountOfWorkers; i++ {
		wg.Add(1)
		go p.worker(&wg)
	}
	wg.Wait()
	close(results)
}

/*
 * 创建任务并加入到协程池中
 */
func allocate(HttpModels []HttpModel) {
	for _, row := range HttpModels {
		jobs <- row
	}
	close(jobs)
}

/*
 * 读取返回结果
 */
func (p *Proxy) result(done chan []Result) {
	var tmp = []Result{}
	for result := range results {
		if result.Job.NeedCache {
			cache_key := "HTTP_" + result.Job.HTTPUniqid
			if err := cache.SetCache(cache_key, result.Data, 600); err != nil {
				log.Print("error", fmt.Sprintf("[error]CacheHTTP set cache:%s", err.Error()))
			}
			p.SetDebug(fmt.Sprintf("Cache Set: %s", cache_key), 1)
		}
		tmp = append(tmp, result)
	}
	done <- tmp
}

// SetDebug 写入debug信息
func (p *Proxy) SetDebug(str string, depth int) {
	if p.Debug != nil {
		if depth == 0 {
			depth = 1
		}

		_, file, line, _ := runtime.Caller(depth)
		path := strings.LastIndexByte(file, '/')
		tmp := string([]byte(file)[path+1:]) + "(line " + strconv.Itoa(line) + "): " + str
		*p.Debug = append(*p.Debug, tmp)
	}
}
