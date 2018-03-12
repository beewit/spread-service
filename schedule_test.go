package main

import (
	"testing"
	"github.com/beewit/spread-service/global"
	"net/http/httptest"
	"strings"
	"github.com/stretchr/testify/assert"
	"github.com/labstack/echo"
	"net/url"
	"github.com/beewit/spread-service/handler"
	"net/http"
	"encoding/json"
	"github.com/beewit/beekit/utils/convert"
)

func TestRedisList(t *testing.T) {
	_, err := global.RD.DelSETKeyValue("spread", "4")
	if err != nil {
		println(err.Error())
	}
	c, _ := global.RD.GetSETCount("spread")

	println(c)
	/*k, _ := global.RD.GetSETRandStringRm("spread")
	println(k)*/
}

func TestArray(t *testing.T) {
	arr := []string{"1sd", "2fgg", "3ggj"}
	b, _ := json.Marshal(arr)
	println(string(b))
	var arrs []string
	json.Unmarshal([]byte("[\"www.baidu.com\",\"www.so.com\",\"www.baidu.com\"]"), &arrs)
	json.Unmarshal(b, &arrs)
	arrs, _ = convert.ToStringArray(string(b))

	println(convert.ToString(arrs))
}

func TestGetTask(t *testing.T) {
	e := echo.New()
	f := url.Values{}
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// 断言
	if assert.NoError(t, handler.GetTask(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		t.Log(rec.Body.String())
	}
}

func TestSaveTask(t *testing.T) {
	e := echo.New()
	f := url.Values{}
	f.Set("urls", "[\"www.baidu2.com\",\"www.so2.com\",\"www.baidu2.com\"]")
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// 断言
	if assert.NoError(t, handler.SaveTask(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		t.Log(rec.Body.String())
	}
}

func TestUpdateCompany(t *testing.T) {
	flog := handler.UpdateCompany(
		map[string]interface{}{"id": "1", "name": "", "tel": "023-98565648", "contacts_mobile": ""},
		map[string]interface{}{"name": "张三", "tel": "023-98565647", "contacts_mobile": "182232"})
	println(flog)
}
