package noderank

import (
	"encoding/json"
	"github.com/triasteam/noderank"
	"reflect"
	"testing"
)

func TestGetRank2(t *testing.T) {
	r := buildData()
	_, actural, _ := noderank.CaculateRank(r, 1, 1)

	if reflect.DeepEqual(len(actural), 1) != true {
		t.Error("expected", 1, "but got ", len(actural))
	}

	acturalStr := actural[0].Attestee
	expected := "D"

	if reflect.DeepEqual(acturalStr, expected) != true {
		t.Error("expected", expected, "but got ", acturalStr)
	}
}

func buildData() []byte {
	ctx := teectx{Attestee: "192.168.2.1", Attester: "192.168.3.1", Score: 0, Address: "XXX",
		Time: "2019-12-11 12:29:29", Nonce: 1, Sign: "cccc"}
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
