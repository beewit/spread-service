package handler

import (
	"github.com/beewit/spread-service/global"
	"github.com/beewit/beekit/utils/convert"
	"fmt"
	"strings"
	"github.com/beewit/beekit/utils"
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
