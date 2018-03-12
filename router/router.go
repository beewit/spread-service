package router

import (
	"github.com/beewit/beekit/utils"
	"github.com/beewit/spread-service/global"
	"github.com/beewit/spread-service/handler"

	"github.com/labstack/echo"

	"fmt"

	"github.com/beewit/beekit/utils/convert"
	"github.com/labstack/echo/middleware"
)

func Start() {
	fmt.Printf("登陆授权系统启动")
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	e.POST("/schedule/task/get", handler.GetTask)
	e.POST("/schedule/task/save", handler.SaveTask)
	e.POST("/schedule/task/failed", handler.FailedTask)
	e.POST("/schedule/task/complete", handler.CompleteTask)
	e.POST("/schedule/task/data/save", handler.SaveDataTask)


	e.GET("/schedule/task/get", handler.GetTask)
	e.GET("/schedule/task/save", handler.SaveTask)
	e.GET("/schedule/task/failed", handler.FailedTask)
	e.GET("/schedule/task/complete", handler.CompleteTask)
	e.GET("/schedule/task/data/save", handler.SaveDataTask)

	utils.Open(global.Host)
	port := ":" + convert.ToString(global.Port)
	e.Logger.Fatal(e.Start(port))
}
