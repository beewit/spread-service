package handler

import (
	"encoding/json"
	"io/ioutil"

	"github.com/beewit/beekit/utils"
	"github.com/beewit/beekit/utils/convert"
	"github.com/beewit/beekit/utils/enum"
	"github.com/beewit/hive/global"
	"github.com/labstack/echo"
)

func readBody(c echo.Context) (map[string]string, error) {
	if c.Request().Body == nil {
		return nil, nil
	}
	body, bErr := ioutil.ReadAll(c.Request().Body)
	if bErr != nil {
		global.Log.Error("读取http body失败，原因：%s", bErr.Error())
		return nil, bErr
	}
	defer c.Request().Body.Close()
	var bm map[string]string
	bErr = json.Unmarshal(body, &bm)
	if bErr != nil {
		global.Log.Error("解析http body失败，原因：%s", bErr.Error())
		return nil, bErr
	}
	return bm, bErr
}

func Filter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		acc, errStr := GetAccountByToken(c)
		if errStr != "" {
			return utils.AuthFail(c, errStr)
		}
		c.Set("account", acc)
		return next(c)
	}
}

func GetAccountByToken(c echo.Context) (*global.Account, string) {
	var token string
	token = c.FormValue("token")
	if token == "" {
		bm, _ := readBody(c)
		if bm != nil {
			token = bm["token"]
		}
	}
	if token == "" {
		return nil, "登陆信息token无效，请重新登陆"
	}

	accMapStr, err := global.RD.GetString(token)
	if err != nil {
		global.Log.Error(err.Error())
		return nil, "登陆信息已失效，请重新登陆"
	}
	if accMapStr == "" {
		global.Log.Error(token + "已失效")
		return nil, "登陆信息已失效，请重新登陆"
	}
	accMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(accMapStr), &accMap)
	if err != nil {
		global.Log.Error(accMapStr + "，error：" + err.Error())
		return nil, "登陆信息已失效，请重新登陆"
	}
	m, err := global.DB.Query("SELECT id,nickname,photo,mobile,status,org_id FROM account WHERE id=? LIMIT 1", accMap["id"])
	if err != nil {
		return nil, "获取用户信息失败"
	}
	if convert.ToString(m[0]["status"]) != enum.NORMAL {
		return nil, "用户已被冻结"
	}
	return global.ToMapAccount(m[0]), ""
}

func GetAccount(c echo.Context) (acc *global.Account, err error) {
	itf := c.Get("account")
	if itf == nil {
		err = utils.AuthFailNull(c)
		return
	}
	acc = global.ToInterfaceAccount(itf)
	if acc == nil {
		err = utils.AuthFailNull(c)
		return
	}
	return
}
