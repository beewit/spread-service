package handler

import (
	"github.com/labstack/echo"
	"github.com/beewit/beekit/utils"
	"github.com/beewit/spread-service/global"
	"encoding/json"
	"github.com/beewit/beekit/utils/convert"
	"io/ioutil"
)

//待抓的任务
var taskName = "SPREAD-TASK"
//进行中的任务
var haveName = "SPREAD-HAVE-TASK"
//失败的任务
var failedName = "SPREAD-FAILED-TASK"
//完成的任务
var completeName = "SPREAD-COMPLETE-TASK"

func checkQueue(key, value string) bool {
	x, _ := global.RD.CheckSETString(key, value)
	return x > 0
}

/**
   获取任务
 */
func GetTask(c echo.Context) error {
	url, _ := global.RD.GetSETRandStringRm(taskName)
	if url == "" {
		return utils.NullData(c)
	}
	//存储到进行中的任务
	global.RD.SetSETString(haveName, url)
	return utils.SuccessNullMsg(c, url)
}

/**
	保存任务
 */
func SaveTask(c echo.Context) error {
	urls := c.FormValue("urls")
	var arrs []string
	err := json.Unmarshal([]byte(urls), &arrs)
	if err != nil {
		return utils.ErrorNull(c, "urls地址错误")
	}
	for _, url := range arrs {
		//排除已完成任务，进行中的任务
		if !checkQueue(taskName, url) && !checkQueue(haveName, url) && !checkQueue(completeName, url) {
			global.RD.SetSETString(taskName, url)
		} else {
			global.Log.Warning("此url的队列已存在：%s", url)
		}
	}
	return utils.SuccessNull(c, "任务保存成功")
}

/**
	完成任务
 */
func CompleteTask(c echo.Context) error {
	urls := c.FormValue("urls")
	var arrs []string
	err := json.Unmarshal([]byte(urls), &arrs)
	if err != nil {
		return utils.ErrorNull(c, "urls地址错误")
	}
	for _, url := range arrs {
		//删除进行中的任务
		global.RD.DelSETKeyValue(taskName, url)
		global.RD.SetSETString(completeName, url)
	}
	return utils.SuccessNull(c, "任务保存成功")
}

/**
	失败任务
 */
func FailedTask(c echo.Context) error {
	urls := c.FormValue("urls")
	var arrs []string
	err := json.Unmarshal([]byte(urls), &arrs)
	if err != nil {
		return utils.ErrorNull(c, "urls地址错误")
	}
	for _, url := range arrs {
		//删除进行中的任务
		global.RD.DelSETKeyValue(taskName, url)
		global.RD.SetSETString(failedName, url)
	}
	return utils.SuccessNull(c, "任务保存成功")
}

/**
	保存采集任务
 */
func SaveDataTask(c echo.Context) error {
	body, bErr := ioutil.ReadAll(c.Request().Body)
	if bErr != nil {
		global.Log.Error("读取http body失败，原因：", bErr.Error())
		return bErr
	}
	defer c.Request().Body.Close()
	global.Log.Info("保存任务清洗数据的结果，请求参数："+string(body))
	var mapArr []map[string]interface{}
	bErr = json.Unmarshal(body, &mapArr)
	if mapArr==nil || len(mapArr)<1{
		return utils.ErrorNull(c, "数据结构错误")
	}
	if len(mapArr) > 0 {
		var flog bool
		var name string
		for _, m := range mapArr {
			name = convert.ToString(m["name"])
			oldMap := GetCompany(name)
			if oldMap == nil {
				flog = SaveCompany(m, c.RealIP())
			} else {
				flog = UpdateCompany(oldMap, m)
			}
			if flog {
				global.Log.Info("y 保存成功【%s】", name)
			} else {
				global.Log.Error("save company error")
			}
		}
		return utils.SuccessNull(c, "保存完成")
	}
	return utils.ErrorNull(c, "保存失败")
}
