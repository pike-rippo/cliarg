package cliarg

import "reflect"

var tagKeys []string = []string{
	"short",
	"long",
	"default",
	"required",
	"help",
}

type fieldInfo struct {
	Index       int
	Type        reflect.Kind
	Required    bool
	DefaultBool bool
}

type parseData struct {
	NameMap map[string]int
	InfoMap map[int]fieldInfo
}

func newParseData() (d parseData) {
	d.NameMap = make(map[string]int)
	d.InfoMap = make(map[int]fieldInfo)
	return d
}

func (d *parseData) AddName(s string, i int) bool {
	_, ok := d.NameMap[s]
	if ok {
		return false
	}
	d.NameMap[s] = i
	return true
}

func (d parseData) GetInfo(s string) (fieldInfo, bool) {
	index, ok := d.NameMap[s]
	if !ok {
		return fieldInfo{}, false
	}
	r, ok := d.InfoMap[index]
	return r, ok
}
