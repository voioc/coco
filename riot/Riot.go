package riot

import (
	"fmt"
	"lemon/app/model"
	log "lemon/lib/log"
	"strconv"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}
)

func GetSearch() *riot.Engine {
	return &searcher
}

func InitEngine() {
	// 初始化
	searcher.Init(types.EngineOpts{
		Using:    3,
		GseDict:  "zh",
		UseStore: true,
		// GseDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/dictionary.txt",
	})

	defer searcher.Close()
}

func Save() error {

	engine := model.GetMySQL()

	// // 同步结构体与数据表
	// if err := engine.Sync(new(model.MacVod)); err != nil {
	// 	log.Fatalf("Fail to sync database: %v\n", err)
	// }

	// type MacVod struct {
	// 	VodId       int64
	// 	VodName     string
	// 	VodActor    string
	// 	VodDirector string
	// 	VodWriter   string
	// }

	result := make([]model.MacVod, 0)

	if err := engine.Limit(10, 0).OrderBy("vod_id asc").Find(&result); err != nil {
		log.Print("error", "insert data to mysql error, the error is: "+err.Error())
	}

	// // 初始化
	// searcher.Init(types.EngineOpts{
	// 	Using:    3,
	// 	GseDict:  "zh",
	// 	UseStore: true,
	// 	// GseDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/dictionary.txt",
	// })

	// defer searcher.Close()

	// text := "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"
	// text1 := "在IMAX影院放映时"
	// text2 := "全片以上下扩展至IMAX 1.9：1的宽高比来呈现"
	searcher := GetSearch()
	// 将文档加入索引，docId 从1开始
	// searcher.Index("1", types.DocData{Content: text})
	// searcher.Index("2", types.DocData{Content: text1}, false)
	// searcher.Index("3", types.DocData{Content: text2}, true)
	for _, row := range result {
		content := fmt.Sprintf("vod_id=%d;vod_name=%s;vod_actor=%s;vod_director=%s;vod_writer=%s;vod_hot=%d", row.VodId, row.VodName, row.VodActor, row.VodDirector, row.VodWriter, 0)
		searcher.Index(strconv.Itoa(int(row.VodId)), types.DocData{Content: content})
	}

	// 等待索引刷新完毕
	searcher.Flush()
	// engine.FlushIndex()

	// // 搜索输出格式见 types.SearchResp 结构体
	// fmt.Println(searcher.Search(types.SearchReq{Text: "三十年"}))

	// Output.SetValue(result, StartTime)
	return nil
} /*
 */
