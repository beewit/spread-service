package handler

import (
	"github.com/beewit/spread-service/global"
	"github.com/beewit/beekit/utils/convert"
	"fmt"
	"strings"
	"github.com/beewit/beekit/utils"
	"github.com/labstack/echo"
	"github.com/beewit/beekit/utils/enum"
	"time"
)

func GetCompany(name string) map[string]interface{} {
	rows, err := global.DB.Query("SELECT * FROM company WHERE name=? limit 1", name)
	if err != nil || len(rows) != 1 {
		return nil
	}
	return rows[0]
}
func GetCompanyById(id int64) map[string]interface{} {
	rows, err := global.DB.Query("SELECT * FROM company WHERE id=? limit 1", id)
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
	acc, errStr := GetAccountByToken(c)
	var accId int64 = 0
	if errStr == "" && acc != nil {
		accId = acc.ID
	}
	province := c.FormValue("province")
	city := c.FormValue("city")
	industry := c.FormValue("industry")
	classify := c.FormValue("classify")
	var where string
	if province != "" {
		province = strings.Replace(province, "回族自治区", "", -1)
		province = strings.Replace(province, "维吾尔自治区", "", -1)
		province = strings.Replace(province, "壮族自治区", "", -1)
		province = strings.Replace(province, "特别行政区", "", -1)
		province = strings.Replace(province, "自治区", "", -1)
		where += fmt.Sprintf(" AND (province='%s' OR province='%s市' OR province='%s省')", province, province, province)
	}
	if city != "" {
		where += fmt.Sprintf("  AND city='%s'", city)
	}
	if classify != "" {
		where += " AND (name LIKE '%" + classify + "%' OR intro LIKE '%" + classify + "%' OR main_business LIKE '%" + classify + "%')"
	} else {
		if industry != "" {
			where += " AND (name LIKE '%" + industry + "%' OR intro LIKE '%" + industry + "%' OR main_business LIKE '%" + industry + "%')"
		}
	}
	pageIndex := utils.GetPageIndex(c.FormValue("pageIndex"))
	pageSize := utils.GetPageSize(c.FormValue("pageSize"))
	beginTime := utils.FormatTime(time.Now().Add(-30 * 24 * time.Hour))
	page, err := global.DB.QueryPage(&utils.PageTable{
		Fields: "company.*",
		Table: "company LEFT JOIN account_del_company del ON del.company_id=company.id AND del.account_id=? " +
			"LEFT JOIN account_import_company import ON import.company_id=company.id AND import.account_id=?",
		Where: "company.status=? AND del.company_id IS NULL AND import.company_id IS NULL AND char_length(contacts_mobile)=11" +
			" AND company.ct_time>? " + where,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Order:     "company.ct_time DESC",
	}, accId, accId, enum.NORMAL, beginTime)
	if err != nil {
		global.Log.Error("QueryPage company sql error:%s", err.Error())
		return utils.ErrorNull(c, "数据异常")
	}
	if page == nil {
		return utils.NullData(c)
	}
	m, _ := global.DB.Query("SELECT COUNT(1) AS count FROM company")
	if len(m) > 0 {
		page.Count = convert.MustInt(m[0]["count"])
	}
	return utils.Success(c, "获取数据成功", page)
}

/**
	批量导入的数据
 */
func ImportCompanyLog(c echo.Context) error {
	acc, err := GetAccount(c)
	if err != nil {
		return err
	}
	idsStr := c.FormValue("ids")
	if idsStr == "" {
		return utils.ErrorNull(c, "无有效id参数")
	}
	var listMap []map[string]interface{}
	ids := strings.Split(idsStr, ",")
	for i := 0; i < len(ids); i++ {
		listMap = append(listMap, map[string]interface{}{
			"id":         utils.ID(),
			"company_id": ids[i],
			"account_id": acc.ID,
			"ct_time":    utils.CurrentTime(),
			"status":     0,
		})
	}
	//添加导入数据记录
	_, err = global.DB.InsertMapList("account_import_company", listMap)
	if err != nil {
		global.Log.Error("ImportCompany account_import_company sql ，error：", err.Error())
		return utils.ErrorNull(c, "添加导入数据记录失败")
	}
	go func() {
		//对数据添加删除数量
		global.DB.Update(fmt.Sprintf("UPDATE company SET import_num=import_num+1 WHERE id IN (%s)", idsStr))
	}()
	return utils.SuccessNull(c, "添加导入数据记录成功")
}

/**
获取导入的数据
 */
func GetImportCompany(c echo.Context) error {
	acc, err := GetAccount(c)
	if err != nil {
		return err
	}
	rows, err := global.DB.Query("SELECT `import`.id as importId,name,contacts_name,contacts_mobile FROM account_import_company import "+
		"LEFT JOIN company ON import.company_id=company.id WHERE import.account_id=? AND import.status=0", acc.ID)
	if err != nil {
		global.Log.Error("GetImportCompany account_import_company sql ，error：", err.Error())
		return utils.ErrorNull(c, "获取导入数据失败")
	}
	if len(rows) < 1 {
		return utils.NullData(c)
	}
	return utils.SuccessNullMsg(c, rows)
}

/**
获取导入的数据
 */
func GetImportCompanyCount(c echo.Context) error {
	acc, err := GetAccount(c)
	if err != nil {
		return err
	}
	rows, err := global.DB.Query("SELECT count(1) as c FROM account_import_company import "+
		"LEFT JOIN company ON import.company_id=company.id WHERE import.account_id=? AND import.status=0", acc.ID)
	if err != nil {
		global.Log.Error("GetImportCompany account_import_company sql ，error：", err.Error())
		return utils.ErrorNull(c, "获取导入数据失败")
	}
	if len(rows) < 1 {
		return utils.SuccessNullMsg(c, 0)
	}
	return utils.SuccessNullMsg(c, rows[0]["c"])
}

/**
修改用户已清理导入数据
 */
func UpdateImportCompanyStatus(c echo.Context) error {
	acc, err := GetAccount(c)
	if err != nil {
		return err
	}
	idsStr := c.FormValue("ids")
	if idsStr == "" {
		return utils.ErrorNull(c, "无有效id参数")
	}
	x, err := global.DB.Update(fmt.Sprintf("UPDATE account_import_company SET status=1 WHERE account_id=? AND id IN (%s)", idsStr), acc.ID)
	if err != nil {
		global.Log.Error("GetImportCompany account_import_company sql ，error：", err.Error())
		return utils.ErrorNull(c, "获取导入数据失败")
	}
	if x < 1 {
		return utils.NullData(c)
	}
	return utils.SuccessNull(c, "修改导入状态成功")
}

/**
 添加删除数据记录
 */
func DelCompanyLog(c echo.Context) error {
	acc, err := GetAccount(c)
	if err != nil {
		return err
	}
	id := c.FormValue("id")
	if id == "" || !utils.IsValidNumber(id) {
		return utils.ErrorNull(c, "无有效id参数")
	}
	m := GetCompanyById(convert.MustInt64(id))
	if m == nil {
		return utils.ErrorNull(c, "无有效数据")
	}
	//添加导入数据记录
	_, err = global.DB.InsertMap("account_del_company", map[string]interface{}{
		"id":         utils.ID(),
		"company_id": id,
		"account_id": acc.ID,
		"ct_time":    utils.CurrentTime(),
	})
	if err != nil {
		global.Log.Error("DelCompany account_del_company sql ，error：", err.Error())
		return utils.ErrorNull(c, "添加删除数据记录失败")
	}
	go func() {
		//对数据添加删除数量
		global.DB.Update("UPDATE company SET del_num=del_num+1 WHERE id=?", id)
	}()
	return utils.SuccessNull(c, "添加删除数据记录成功")
}
