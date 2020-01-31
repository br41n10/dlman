package data

import (
	"database/sql"
	"github.com/labstack/gommon/log"
	"net/url"
)

type taskStatus string

const (
	RedisDownloadTaskQueueId = "dlman_task_queue"
	RedisMirrorTaskQueueId   = "dlman_mirror_task_queue"
)

const (
	taskNotStarted    taskStatus = "NOT_STARTED"
	taskInQueue       taskStatus = "IN_QUEUE"
	taskDownloading   taskStatus = "DOWNLOADING"
	TaskDownloaded    taskStatus = "DOWNLOADED"
	TaskTransToMirror taskStatus = "TRANS_TO_MIRROR"
	TaskFinished      taskStatus = "FINISHED"

	TaskDownloadError taskStatus = "DOWNLOAD_ERROR"
)

// 任务结构体
type Task struct {
	Id               int64
	userId           int64
	OriginalUrl      *url.URL // 任务包括url信息
	FinalUrl         *url.URL
	File             *File
	Status           taskStatus
	downloadProgress float32
}

// 初始化一个task
func NewTask(originUrl *url.URL) (*Task, error) {

	// 初始化文件对象
	file, err := NewFile()
	if err != nil {
		return nil, err
	}

	t := Task{
		userId:           1,
		OriginalUrl:      originUrl,
		File:             file,
		Status:           taskNotStarted,
		downloadProgress: 0,
	}

	// 写到数据库中
	result, err := Db.Exec(
		"insert into Task (file_id, download_progress, user_id, origin_url, status) values (?, ?, ?, ?, ?)",
		t.File.Id, t.downloadProgress, t.userId, t.OriginalUrl.String(), t.Status)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	t.Id = id

	return &t, nil
}

// 根据id获取task
// 从数据库中查询到task后，需要再根据file id把file获取到
func GetTaskById(id int64) (*Task, error) {
	task := &Task{}

	// 中间变量
	var (
		fileId         int64
		rawOriginalUrl sql.NullString
		rawFinalUrl    sql.NullString
	)

	row := Db.QueryRow("select id, file_id, download_progress, user_id, origin_url, final_url, status from Task where id = ?", id)
	err := row.Scan(&task.Id, &fileId, &task.downloadProgress, &task.userId, &rawOriginalUrl, &rawFinalUrl, &task.Status)
	if err != nil {
		log.Errorf("GetTaskById查询错误，id：%d", id)
		return nil, err
	}

	file, err := GetFileById(fileId)
	if err != nil {
		log.Errorf("GetTaskById|GetFileById查询错误，fileId: %d", fileId)
		return nil, err
	}
	task.OriginalUrl, _ = url.Parse(rawOriginalUrl.String) // 这两个不会出错，在入库的时候是URL.String()的
	task.FinalUrl, _ = url.Parse(rawFinalUrl.String)
	task.File = file

	return task, nil
}

// 设置 Task 的 status
func (t *Task) SetStatus(ts taskStatus) error {
	_, err := Db.Exec("update Task set status = ? where id = ?", ts, t.Id)
	if err != nil {
		log.Errorf("Task|SetStatus|id: %d|status: %s", t.Id, ts)
		return err
	}
	return nil
}

// 将 Task 添加到任务队列中
func (t *Task) Start() error {
	err := TaskDownloadEnqueue(t.Id)
	if err != nil {
		log.Errorf("Task Start|FAIL|id: %d|err: %v", t.Id, err)
		return err
	}

	// 修改数据库中 Task 状态
	// TODO: 此处可能会造成状态不一致
	err = t.SetStatus(taskInQueue)
	if err != nil {
		return err
	}

	return nil
}

func TaskDownloadEnqueue(taskId int64) error {
	err := RedisCli.RPush(RedisDownloadTaskQueueId, taskId).Err()
	return err
}

func TaskMirrorEnqueue(taskId int64) error {
	err := RedisCli.RPush(RedisMirrorTaskQueueId, taskId).Err()
	return err
}

func (t *Task) SetFinalUrl(u *url.URL) error {
	t.FinalUrl = u
	_, err := Db.Exec("update Task set final_url = ? where id = ?", u.String(), t.Id)
	if err != nil {
		log.Errorf("Task|SetFinalUrl|id: %d|FinalUrl: %s", t.Id, u.String())
		return err
	}
	return nil
}

func (t *Task) SetFilename(fn string) error {
	t.File.Name = sql.NullString{
		String: fn,
		Valid:  true,
	}

	_, err := Db.Exec("update File set name = ? where id = ?", t.File.Name, t.File.Id)
	if err != nil {
		log.Errorf("Task|SetFilename|id: %d|Filename: %s", t.Id, t.File.Name.String)
		return err
	}
	return nil
}
