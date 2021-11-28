package main

import (
	"encoding/json"
	"fmt"
)

type jsonArr []interface{}

type jsonObj map[string]interface{}

//
type CmdResp struct {
	Cmds jsonObj `json:"commands"`
}

//
func getType(i interface{}) string {
	switch v := i.(type) {
	case string:
		return fmt.Sprintf("str: '%+v'", v)
	case int, int16, int32, int64, uint, uint64, uint16, uint32, float32, float64:
		return fmt.Sprintf("num: %v", v)
	case jsonObj:
		return v.String()
	case jsonArr:
		return v.String()
	case CmdResp:
		return v.String()
	}
	if o, ok := i.(jsonArr); ok {
		return o.String()
	}
	if o, ok := i.(jsonObj); ok {
		return o.String()
	}
	if o, ok := i.([]interface{}); ok {
		r := "["
		l := len(o)
		for k, v := range o {
			r = r + fmt.Sprintf("%v", getType(v))
			if k < l-1 {
				r = r + ", "
			}
		}
		r = r + "]"
		return r
	}
	if o, ok := i.(map[string]interface{}); ok {
		r := "{"
		n := 1
		l := len(o)
		for _, v := range o {
			r = r + fmt.Sprintf("%v", getType(v))
			if n < l {
				r = r + ", "
			}
			n++
		}
		r = r + "}"
		return r
	}
	return fmt.Sprintf("unknown %T -> %v", i, i)
}

//
func (j jsonObj) String() string {
	ret := "jObj:{"
	for _, v := range j {
		// At this point the v loose the  knowledge of type, and we have to search by v.(specific type)
		ret = ret + getType(v)
	}
	ret = ret + "}"
	return ret
}

//
func (j jsonArr) String() string {
	ret := "jArr:["
	for _, v := range j {
		// At this point the v loose the  knowledge of type, and we have to search by v.(specific type)
		ret = ret + getType(v)
	}
	ret = ret + "]"
	return ret
}

//
func (j CmdResp) String() string {
	ret := "Cmds:{"
	ret = ret + getType(j.Cmds)
	ret = ret + "}"
	return ret
}

//
func sample_main() {
	v := CmdResp{
		Cmds: jsonObj{
			"database": "person",
			"info": jsonObj{
				"name":    "person-0",
				"age":     23,
				"parents": jsonArr{"person1", "person2"},
			},
		},
	}
	if data, e := json.Marshal(v); e != nil {
		fmt.Println("error processing marshal, ", e)
	} else {
		var vv CmdResp
		fmt.Println("encoded string is ", data)
		if e = json.Unmarshal(data, &vv); e != nil {
			fmt.Println("failed to unmarshall data", e)
		} else {
			fmt.Println(fmt.Sprintf("Got back data '%v'", getType(vv)))
		}
	}

}
