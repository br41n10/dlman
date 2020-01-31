package main

import (
	"dlman/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// 放一些http框架相关的

// 任务目前很简单，只是记录时间、邮箱、url、fileId，主要用作通知用

// 对于提交文件的url，判断有没有存在于file表中
////// 对于存在于file中的url，head一下， 获取meta信息，和表中的记录做比对
////////// 如果是一样的，则直接返回该文件地址
////////// 如果不一样，新的下载-->上传到oss-->通知用户邮箱
////// 对于file表中没有的url，新的下载-->上传到oss-->通知用户邮箱

// 先做成全部重新下载的
func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.DEBUG)

	// task
	e.POST("/tasks", handler.CreateTask)
	e.GET("/tasks/:id", handler.GetTask)
	e.GET("/tasks/:taskId/file", handler.GetTaskFile)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}
