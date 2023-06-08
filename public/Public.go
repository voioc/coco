package public

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/voioc/coco/logzap"
)

func InSlice(obj interface{}, target []string) bool {
	for _, val := range target {
		if obj == val {
			return true
		}
	}

	return false
}

//ip 转 int64
func IP2Long(ip string) (int, error) {
	ret := big.NewInt(0)
	flag := net.ParseIP(ip).To4()

	if flag == nil {
		return 0, errors.New("the ip is illegal")
	}

	ret.SetBytes(flag)
	return int(ret.Int64()), nil
}

// int64 转 ip
func Long2IP(ip int64) net.IP {
	var tmp [4]byte
	tmp[0] = byte(ip & 0xFF)
	tmp[1] = byte((ip >> 8) & 0xFF)
	tmp[2] = byte((ip >> 16) & 0xFF)
	tmp[3] = byte((ip >> 24) & 0xFF)

	return net.IPv4(tmp[3], tmp[2], tmp[1], tmp[0])
}

var cost map[string]time.Time

// var rw sync.RWMutex

func Flagtime(flag string) string {
	if cost == nil {
		cost = map[string]time.Time{}
	}

	if _, ok := cost[flag]; !ok {
		cost[flag] = time.Now()
	}

	tc := time.Since(cost[flag])
	delete(cost, flag)

	return fmt.Sprintf("%vms", tc)
}

func TimeCost(start time.Time) string {
	tc := time.Since(start)
	return fmt.Sprintf("%v", tc)
}

// 检测文件是否存在
func File_exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

//复制文件
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// GetRoot 获取程序根目录
func GetRoot() string {
	dir, err := filepath.Abs(filepath.Dir("")) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}

	dir += "/"                                 // 以/结尾
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

// TrimHTML 清除html标签
func TrimHTML(src string) string {
	src = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(src, "&nbsp;", ""), "\r\n", ""), " ", "")

	//将HTML标签全转换成小写
	re, _ := regexp.Compile(`\\<[\\S\\s]+?\\>`)
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile(`\\<style[\\S\\s]+?\\</style\\>`)
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile(`\\<script[\\S\\s]+?\\</script\\>`)
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile(`\\<[\\S\\s]+?\\>`)
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile(`\\s{2,}`)
	src = re.ReplaceAllString(src, "\n")

	//去除<p>
	re, _ = regexp.Compile(`<.*?>`)
	src = re.ReplaceAllString(src, "")
	return strings.TrimSpace(src)
}

// FilePutContents 写入文件信息
func FilePutContents(file string, content interface{}, isShowTime bool) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logzap.Dx(context.Background(), "FilePutContents", "file: %s | error: %s", file, err.Error())
	}

	c := ""
	if v, ok := content.([]byte); ok {
		c = string(v)
	} else if v, ok := content.(string); ok {
		c = v
	}

	if isShowTime {
		now := "[" + time.Now().Local().Format("2006-01-02 15:04:05") + "] "
		c = now + c
	}

	if _, err := f.WriteString(c + "\n"); err != nil {
		logzap.Dx(context.Background(), "FilePutContents", "file: %s | error: %s", file, err.Error())
	}
}

// float 类型精度问题
func ChangeNumber(f float64, m int) string {
	n := strconv.FormatFloat(f, 'f', -1, 32)
	if n == "" {
		return ""
	}

	if m >= len(n) {
		return n
	}

	newn := strings.Split(n, ".")
	if len(newn) < 2 || m >= len(newn[1]) {
		return n
	}

	return newn[0] + "." + newn[1][:m]
}

func FormatError(msg, err string, debug interface{}) string {
	errInfo := map[string]interface{}{
		"msg":   msg,
		"debug": debug,
		"error": err,
	}

	str, _ := jsoniter.MarshalToString(errInfo)

	return FormatJson(str)
}

func FormatJson(text string) string {
	newText := strings.ReplaceAll(text, "\"{", "{")
	newText = strings.ReplaceAll(newText, "}\"", "}")
	newText = strings.ReplaceAll(newText, "\\\"", "\"")
	return newText
}

func RuneToString(data string) string {
	result := ""
	for _, row := range data {
		if result == "" {
			result = strconv.Itoa(int(row))
		} else {
			result += "_" + strconv.Itoa(int(row))
		}
	}

	return result
}

func UrlEncode(params map[string]string) string {
	encode := url.Values{}
	for k, v := range params {
		encode.Add(k, v)
	}
	//   params.Add("name", "中国")
	//   params.Add("phone", "+8613000000000")
	return encode.Encode()
}
