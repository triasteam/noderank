package noderank

import (
	"github.com/wunder3605/noderank"
	"reflect"
	"testing"
)

func TestGetRank(t *testing.T) {

	info := []string{"A", "B", "1"}
	noderank.AddAttestationInfo("","",info)
	info = []string{"B", "C", "1"}
	noderank.AddAttestationInfo("","",info)
	info = []string{"C", "D", "1"}
	noderank.AddAttestationInfo("","",info)
	info = []string{"D", "A", "1"}
	noderank.AddAttestationInfo("","",info)
	info = []string{"A", "C", "1"}
	noderank.AddAttestationInfo("","",info)

	_,actural,_ := noderank.GetRank("",1, 1)

	if reflect.DeepEqual(len(actural), 1) != true {
		t.Error("expected", 1, "but got ", len(actural))
	}

	acturalStr := actural[0].Attestee
	expected := "D"

	if reflect.DeepEqual(acturalStr, expected) != true {
		t.Error("expected", expected, "but got ", acturalStr)
	}
}
