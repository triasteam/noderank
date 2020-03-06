package noderank

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestAddAttestationInfo(t *testing.T) {
	info := make([]string, 5)
	info[0] = "10.10.1.1"
	info[1] = "10.10.1.2"
	info[2] = "1"
	info[3] = "1"
	info[4] = "2019-01-01"

	err1 := AddAttestationInfo("", "http://localhost:14700", info)
	fmt.Println(err1)
}

func TestGetRank2(t *testing.T) {
	r := buildData()
	_, actural, _ := CaculateRank(r, 1, 1)

	if reflect.DeepEqual(len(actural), 1) != true {
		t.Error("expected", 1, "but got ", len(actural))
	}

	//acturalStr := actural[0].Attestee
	//expected := "D"
	//
	//if reflect.DeepEqual(acturalStr, expected) != true {
	//	t.Error("expected", expected, "but got ", acturalStr)
	//}
}

func buildData() []byte {
	ctx := teectx{Attestee: "192.168.2.1", Attester: "192.168.3.1", Score: 0,
		Time: "2019-12-11 12:29:29", Nonce: 1}
	msg := message{TeeNum: 1, TeeContent: []teectx{ctx}}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	msgArr := []string{string(msgBytes[:])}

	msgArrStr, err := json.Marshal(msgArr)

	block := string(msgArrStr[:])

	response := Response{Blocks: block, Duration: 1}

	responseStr, err := json.Marshal(response)

	return responseStr
}
