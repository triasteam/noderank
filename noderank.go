// Copyright by StreamNet team
// 功能描述：
// 1. 拼装新增证实交易请求；
// 2. 获取被证实节点的排名：将dag中的证实交易按一定顺序构造全拓扑序，以分页的方式获取指定n个证实交易作为输入，使用pagerank算法计算出这些节点的排名。

package noderank

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"github.com/triasteam/pagerank"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	url2 "net/url"
	"sort"
	"strconv"
	"time"
)

type Response struct {
	Blocks   string `json:"blocks"`
	Duration int    `json:"duration"`
}

type message struct {
	TeeNum     int64    `json:"tee_num"`
	TeeContent []teectx `json:"tee_content"`
}

type teectx struct {
	Attester string  `json:"attester"`
	Attestee string  `json:"attestee"`
	Score    float64 `json:"score"`
	Address  string  `json:"address,omitempty"`
	Time     string  `json:"time,omitempty"`
	Nonce    int64   `json:"nonce,omitempty"`
	Sign     string  `json:"sign,omitempty"`
}

type teescore struct {
	Attestee string  `json:"attestee"`
	Score    float64 `json:"score"`
}

type teescoreslice []teescore

var url = "http://localhost:14700"
var addr = "JVSVAFSXWHUIZPFDLORNDMASGNXWFGZFMXGLCJQGFWFEZWWOA9KYSPHCLZHFBCOHMNCCBAGNACPIGHVYX"

var (
	file = flag.String("file", "noderank/config.yaml", "IOTA CONFIGURATION")
)

func AddAttestationInfo(addr1 string, url string, info []string) error {
	raw := new(teectx)
	raw.Attester = info[0]
	raw.Attestee = info[1]
	raw.Address = info[3]
	raw.Nonce, _ = strconv.ParseInt(info[4], 10, 64)
	raw.Time = info[5]
	raw.Sign = info[6]
	num, err := strconv.ParseInt(info[2], 10, 64)
	if err != nil {
		return err
	}
	raw.Score = float64(num)
	m := new(message)
	m.TeeNum = 1
	m.TeeContent = []teectx{*raw}
	ms, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if addr1 == "" {
		addr1 = addr
	}

	d := time.Now()
	ds := d.Format("20060102")
	data := "{\"command\":\"storeMessage\",\"address\":" + addr1 + ",\"message\":" + url2.QueryEscape(string(ms[:])) + ",\"tag\":\"" + ds + "TEE\"}"
	_, err = doPost(url, []byte(data))
	if err != nil {
		return err
	}
	return nil
}

//根据条件获取到的所有证实交易都会参与noderank计算，结果返回前numRank的得分和对应得证实交易数（由于待输出数据是以attestee作为key得map中保存，
//   最终输出得teectx最终数量会少于实际）。
//uri StreamNet服务restful地址， peroid表示把全部交易按每页100个分页后所取页数，numRank 取排名后前numRank个被证实节点
func GetRank(uri string, period int64, numRank int64) ([]teescore, []teectx, error) {
	data := "{\"command\":\"getBlocksInPeriodStatement\",\"period\":" + strconv.FormatInt(period, 10) + "}"
	r, err := doPost(uri, []byte(data))
	if err != nil {
		fmt.Println("do post error, data = ", data)
		panic(err)
	}
	return CaculateRank(r, period, numRank)
}

func CaculateRank(r []byte, period int64, numRank int64) ([]teescore, []teectx, error) {
	var result Response
	err := json.Unmarshal(r, &result)
	if err != nil {
		fmt.Println("unmarshal Response error, r = ", r)
		return nil, nil, err
	}

	var msgArr []string
	err = json.Unmarshal([]byte(result.Blocks), &msgArr)
	if err != nil {
		fmt.Println("unmarshal string array error, result.Blocks = ", result.Blocks)
		return nil, nil, err
	}

	graph := pagerank.NewGraph()

	cm := make(map[string]teectx)

	rArr0 := []teectx{}

	for _, m2 := range msgArr {
		msgT, err := url2.QueryUnescape(m2)
		if err != nil {
			fmt.Println("QueryUnescape error, m2 = ", m2)
			return nil, nil, err
		}
		var msg message
		err = json.Unmarshal([]byte(msgT), &msg)
		if err != nil {
			fmt.Println("unmarshal message error, msgT = ", msgT)
			return nil, nil, err
		}

		rArr := msg.TeeContent

		for _, r := range rArr {
			if math.IsNaN(r.Score) || math.IsInf(r.Score, 0) {
				fmt.Println("un invalid rank param. score : ", r.Score)
			} else {
				if r.Score == 0 {
					fmt.Println("un invalid rank param. score is zero.")
				}
				graph.Link(r.Attester, r.Attestee, r.Score)
				cm[r.Attestee] = teectx{r.Attester, r.Attestee, r.Score, "", "", 0, ""}
				rArr0 = append(rArr0, r)
			}
		}
	}

	var rst []teescore
	var teectxslice []teectx

	graph.Rank(0.85, 0.0001, func(attestee string, score float64) {
		tee := teescore{attestee, FloatRound(score, 8)}
		rst = append(rst, tee)
	})
	sort.Sort(teescoreslice(rst))
	if len(rst) < 1 {
		return nil, nil, nil
	}

	endIdx := int64(len(rst))
	if endIdx > numRank {
		endIdx = numRank
	}

	rst = rst[0:endIdx]
	//for _, r := range rst {
	//	if v, ok := cm[r.Attestee]; ok {
	//		teectxslice = append(teectxslice, v)
	//	}
	//}
	//以结果的Attestee作为key
	scoreMap := make(map[string]float64)
	for _, r := range rst {
		scoreMap[r.Attestee] = r.Score
	}

	//遍历数组，获取前n个排名的被实节点对应的证实交易。
	for _, r := range rArr0 {
		if scoreMap[r.Attestee] != 0 {
			teectxslice = append(teectxslice, r)
		}
	}

	return rst, teectxslice, nil
}

func FloatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

func PrintHCGraph(uri string, period string) error {
	data := "{\"command\":\"getBlocksInPeriodStatement\",\"period\":" + period + "}"
	r, err := doPost(uri, []byte(data))
	if err != nil {
		return err
	}
	var result Response
	err = json.Unmarshal(r, &result)
	if err != nil {
		fmt.Println(r)
	}
	fmt.Println(result.Duration)
	fmt.Println(result.Blocks)

	var msgArr []string
	err = json.Unmarshal([]byte(result.Blocks), &msgArr)
	if err != nil {
		log.Panic(err)
	}

	graph := gographviz.NewGraph()

	for _, m2 := range msgArr {
		msgT, err := url2.QueryUnescape(m2)
		if err != nil {
			log.Panicln(err)
		}
		fmt.Println("message : " + msgT)
		var msg message
		err = json.Unmarshal([]byte(msgT), &msg)
		if err != nil {
			log.Panic(err)
		}

		rArr := msg.TeeContent
		for _, r := range rArr {
			//score := strconv.FormatUint(uint64(r.Score), 10) // TODO add this score info
			graph.AddNode("G", r.Attestee, nil)
			graph.AddNode("G", r.Attester, nil)
			graph.AddEdge(r.Attester, r.Attestee, true, nil)
			if err != nil {
				log.Panic(err)
			}
		}
	}

	output := graph.String()
	fmt.Println(output)
	return nil
}

func doPost(uri string, d []byte) ([]byte, error) {
	if uri == "" {
		uri = url
	}
	fmt.Println("node rank request iota url is: ", uri)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(d))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-IOTA-API-Version", "1")

	client := &http.Client{}
	res, err := client.Do(req)
	fmt.Println("request result:", res, ", err:", err)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	r, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r teescoreslice) Len() int {
	return len(r)
}

func (r teescoreslice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r teescoreslice) Less(i, j int) bool {
	return r[j].Score < r[i].Score
}
