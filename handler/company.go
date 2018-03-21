package handler

import (
	"github.com/beewit/spread-service/global"
	"github.com/beewit/beekit/utils/convert"
	"fmt"
	"strings"
	"github.com/beewit/beekit/utils"
	"github.com/labstack/echo"
	"github.com/beewit/beekit/utils/enum"
)

func GetCompany(name string) map[string]interface{} {
	rows, err := global.DB.Query("SELECT * FROM company WHERE name=? limit 1", name)
	if err != nil || len(rows) != 1 {
		return nil
	}
	return rows[0]
}

func SaveCompany(m map[string]interface{}, ip string) bool {
	if m == nil {
		return false
	}
	m["id"] = utils.ID()
	m["ct_time"] = utils.CurrentTime()
	m["ct_ip"] = ip
	_, err := global.DB.InsertMap("company", m)
	if err != nil {
		global.Log.Error("saveCompany sql error：%s", err.Error())
		return false
	}
	return true
}

func UpdateCompany(m map[string]interface{}, newMap map[string]interface{}) bool {
	if m == nil || newMap == nil {
		return false
	}
	var setKey []string
	var setVal []interface{}
	var nv interface{}
	for k, v := range m {
		nv = newMap[k]
		if convert.ToString(v) == "" && convert.ToString(nv) != "" {
			setKey = append(setKey, k)
			setVal = append(setVal, nv)
		}
	}
	if len(setKey) < 1 && len(setVal) < 1 {
		return false
	}
	sql := fmt.Sprintf("UPDATE company SET %s WHERE id=?", strings.Join(setKey, "=?,")+"=?")
	setVal = append(setVal, m["id"])
	global.Log.Info(sql)
	global.Log.Info(convert.ToString(setVal))
	_, err := global.DB.Update(sql, setVal...)
	if err != nil {
		global.Log.Error(err.Error())
		return false
	}
	return true
}

func GetCompanyPage(c echo.Context) error {
	_, err := GetAccount(c)
	if err != nil {
		return utils.AuthFailNull(c)
	}
	province := c.FormValue("province")
	city := c.FormValue("city")
	//未过期并未领完的
	var where string
	if province != "" {
		where += fmt.Sprintf(" AND (province='%s市')", province)
	}
	if city != "" {
		where += fmt.Sprintf("  AND city='%s'", city)
	}
	pageIndex := utils.GetPageIndex(c.FormValue("pageIndex"))
	pageSize := utils.GetPageSize(c.FormValue("pageSize"))
	page, err := global.DB.QueryPage(&utils.PageTable{
		Fields:    "*",
		Table:     "company",
		Where:     " status=?" + where,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Order:     "ct_time DESC",
	}, enum.NORMAL)
	if err != nil {
		global.Log.Error("QueryPage company sql error:%s", err.Error())
		return utils.ErrorNull(c, "数据异常")
	}
	if page == nil {
		return utils.NullData(c)
	}
	return utils.Success(c, "获取数据成功", page)
}
