package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	strings2 "megin/library/strings"
	"megin/system"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
)

func Post(path string, params interface{}) *httptest.ResponseRecorder {
	r := system.SetupTestRouter()
	body, _ := json.Marshal(params)
	fmt.Println("请求Path:", path)
	//fmt.Println("请求参数:", string(body))
	req := httptest.NewRequest("POST", path, bytes.NewReader(body))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Body.String()
	fmt.Println("返回数据:", resp)
	return w
}

//	r := system.SetupTestRouter()
//	req := httptest.NewRequest("GET", path+"?"+querystring, strings.NewReader(""))
//	//json格式提交
//	req.Header.Add("Content-type", "application/json;charset=utf-8")
//	req.Header.Add("platform", "PH")
//	w := httptest.NewRecorder()
//	r.ServeHTTP(w, req)
//	return w

func Get(path string, params any) *httptest.ResponseRecorder {
	r := system.SetupTestRouter()
	querystring, err := strings2.HttpBuildQuery(params)
	if err != nil {
		panic("参数不正确:" + err.Error())
	}

	req := httptest.NewRequest("GET", path+"?"+querystring, strings.NewReader(querystring))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Delete(path string) *httptest.ResponseRecorder {
	r := system.SetupTestRouter()
	req := httptest.NewRequest("DELETE", path, strings.NewReader(""))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Echo(data interface{}) {
	t := reflect.TypeOf(data).String()
	var res []byte
	if t == "string" {
		res = []byte(cast.ToString(data))
	} else {
		res, _ = json.Marshal(data)
	}

	var out bytes.Buffer
	_ = json.Indent(&out, res, "", "\t")
	out.WriteTo(os.Stdout)
	fmt.Printf("\n")
}
