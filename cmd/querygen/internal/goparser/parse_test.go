package goparser

import (
	"go/importer"
	"testing"
)

var testPath = "../../../../x/twap/client/queryproto/query.pb.go"
var testdirPath = "../../../../x/twap/client/queryproto/"

func TestParse(t *testing.T) {
	p := Parser{importer.Default()}
	p.Parse(testPath, nil)
	t.FailNow()
}
