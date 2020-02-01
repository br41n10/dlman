package handler

import (
	"dlman/data"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/url"
	"strconv"
)

type commonResp struct {
	Code code        `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type createTaskResp struct {
	Id int64 `json:"id"`
}

type getTaskResp struct {
	Id          int64  `json:"id"`
	OriginalUrl string `json:"original_url"`
	FinalUrl    string `json:"final_url"`
	Status      string `json:"status"`

	// file
	FileId       int64  `json:"file_id"`
	FileName     string `json:"file_name"`
	FileVersion  string `json:"file_version"`
	MirrorUrl    string `json:"mirror_url"`
	DownloadTime int64  `json:"download_time"`
}

func CreateTask(c echo.Context) error {

	rawOriginalUrl := c.FormValue("original_url")
	c.Logger().Infof("original url for new task: %s", rawOriginalUrl)

	if rawOriginalUrl == "" {
		return c.JSON(http.StatusBadRequest, nil)
	}

	originalUrl, err := url.Parse(rawOriginalUrl)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// 创建 task
	task, err := data.NewTask(originalUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 启动 task
	err = task.Start()
	if err != nil {
		// 启动失败，等待回扫机制将该 task 重新加入队列
		c.Logger().Errorf("task start error, id: %d", task.Id)
	}

	resp := createTaskResp{
		Id: task.Id,
	}

	return commonJSON(c, http.StatusOK, codeOK, "请求成功", resp)
}

func GetTask(c echo.Context) error {
	taskIdStr := c.Param("id")
	if len(taskIdStr) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}

	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	// 获取 task
	task, err := data.GetTaskById(taskId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 根绝 task 的情况，完善 file 信息
	var (
		fileId        int64
		fileName      string
		fileVersion   string
		mirrorUrl     string
		downloadTimes int64
	)
	if task.Status == data.TaskFinished {
		fileId = task.File.Id
		fileName = task.File.Name.String
		fileVersion = task.File.Version.String
		mirrorUrl = task.File.Mirror.Url.String()
		downloadTimes = task.File.DownloadTime
	}

	resp := getTaskResp{
		// task
		Id:          task.Id,
		OriginalUrl: task.OriginalUrl.String(),
		FinalUrl:    task.FinalUrl.String(),
		Status:      string(task.Status),
		// file
		FileId:       fileId,
		FileName:     fileName,
		FileVersion:  fileVersion,
		MirrorUrl:    mirrorUrl,
		DownloadTime: downloadTimes,
	}

	return commonJSON(c, http.StatusOK, codeOK, "请求成功", resp)
}

// 下载文件接口，根据文件的 carrier 生成正确的下载链接
// 先鉴权，看这个人是否可以请求资源，做法是拉取用户 tsak 表中 file id 列
// 看请求的文件在不在其中。TODO: 这里有点蠢，等碰到性能问题，再考虑做缓存。
func GetTaskFile(c echo.Context) error {
	taskIdStr := c.Param("taskId")
	if len(taskIdStr) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}

	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	task, err := data.GetTaskById(taskId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if task.Status != data.TaskFinished {
		// 任务还没完成，说明他是直接调用的接口，TODO: 封号
		return c.String(http.StatusForbidden, "不要乱搞我的网站哦")
	}

	// 根据不同的 carrier 获取正确的链接返回给用户
	if task.File.Mirror.Carrier == data.CarrierLocal {
		// 本地文件
		return c.Attachment(task.File.LocalAbsPath(), task.File.Name.String)
	} else {
		log.Errorf("GetTaskFile | carrier unknown: %s", task.File.Mirror.Carrier)
	}
	return nil
}
