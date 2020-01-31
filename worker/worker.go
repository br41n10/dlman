// worker 包从redis队列中获取任务，然后执行
// TODO: 有点问题，如果downloader在运行过程中被关闭，那么这个任务就没法很好的报告出去
// TODO: 如果有的 task 已经从队列中取出，但是还没有被阻塞在 taskIds <- taskId的时候，程序关闭，则该task已经被取出，但是没有被运行
// TODO: trans to mirror 还没写完，不影响已有的代码

package main

import (
	"dlman/config"
	"dlman/data"
	"github.com/labstack/gommon/log"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	downloaderCount   int
	mirrorWorkerCount int
)

func init() {
	downloaderCount = 2
	mirrorWorkerCount = 1
}

type taskResult struct {
	task   *data.Task
	taskId int64
	done   bool
}

type mirrorResult struct {
	task   *data.Task
	taskId int64
	done   bool
}

// 下载的 worker
func downloader(id int, taskIdChan <-chan int64, resultChan chan<- taskResult) {

	for true {

		log.Infof("downloader %d | waiting for task", id)

		// 接收 task
		taskId := <-taskIdChan
		log.Infof("downloader %d | received task: %d", id, taskId)

		// 干活

		// 获取 task
		t, err := data.GetTaskById(taskId)
		if err != nil {
			result := taskResult{taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		// 根据task中创建好的本地文件打开该文件
		filePath := t.File.LocalAbsPath()
		outFile, err := os.OpenFile(filePath, os.O_RDWR, 666)

		if err != nil {
			outFile.Close()
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		// 请求 task 中指定的远程资源
		resp, err := http.Get(t.OriginalUrl.String())
		if err != nil {
			outFile.Close()
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		// 根据请求的内容，设置task的一些信息
		// 更新 final url
		err = t.SetFinalUrl(resp.Request.URL)
		if err != nil {
			outFile.Close()
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}
		// 更新 file name
		filename := getFilename(resp, t.OriginalUrl)
		err = t.SetFilename(filename)
		if err != nil {
			outFile.Close()
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		// TODO: 更新 version

		// 将响应体写入上面打开的 outFile 中
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			outFile.Close()
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		// 关闭文件
		err = outFile.Close()
		if err != nil {
			result := taskResult{task: t, taskId: taskId, done: false}
			resultChan <- result
			continue
		}

		log.Infof("downloader %d | task done: %d", id, taskId)
		result := taskResult{task: t, taskId: taskId, done: true}
		resultChan <- result
	}
}

func mirrorWorker(id int, mirrorTaskIdChan <-chan int64, mirrorResultChan chan<- mirrorResult) {
	// 在这里选择 mirror 的目的地
	// 目前只有 LOCAL
	for true {

		log.Infof("mirror worker %d | waiting for task", id)

		// 接收 task
		taskId := <-mirrorTaskIdChan
		log.Infof("mirror worker %d | received task: %d", id, taskId)

		// 干活

		// 获取 task
		t, err := data.GetTaskById(taskId)
		if err != nil {
			result := mirrorResult{taskId: taskId, done: false}
			mirrorResultChan <- result
			continue
		}

		if config.MirrorCarrier == string(data.CarrierLocal) {
			// do nothing trans file to another place

			// 修改 数据库中 的 carrier
			err := t.File.SetCarrier(data.CarrierLocal)
			if err != nil {
				log.Errorf("mirror worker %d ERROR | task: %d", id, taskId)
				result := mirrorResult{task: t, taskId: taskId, done: false}
				mirrorResultChan <- result
				continue
			}
			// 正常
			log.Infof("mirror worker %d | task done: %d", id, taskId)
			result := mirrorResult{task: t, taskId: t.Id, done: true}
			mirrorResultChan <- result
		} else {
			log.Errorf("mirror worker %d | unknown mirror carrier: %s", t.Id, config.MirrorCarrier)
		}
	}
}

// TODO: 程序运行前先判断redis中有没有那个key

func main() {

	// download worker starting...
	taskIdChan := make(chan int64)
	resultChan := make(chan taskResult) // 做成结构体，包含id和结果
	for w := 1; w <= downloaderCount; w++ {
		go downloader(w, taskIdChan, resultChan)
	}

	// mirror worker starting...
	mirrorTaskIdChan := make(chan int64)
	mirrorResultChan := make(chan mirrorResult) // 做成结构体，包含id和结果
	for w := 1; w <= mirrorWorkerCount; w++ {
		go mirrorWorker(w, mirrorTaskIdChan, mirrorResultChan)
	}

	// download dispatcher
	// 从 redis 队列中取出要下载的 task id
	// 没有用BLPop是因为一次会把所有的都返回
	go func() {
		for true {
			time.Sleep(time.Second)

			// 取出 task
			taskIdStr := data.RedisCli.LPop(data.RedisDownloadTaskQueueId).Val()

			if len(taskIdStr) == 0 {
				// 当前redis队列为空
				continue
			}
			taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
			if err != nil {
				log.Errorf("download dispatcher | task from queue error: %s", taskIdStr)
				continue
			}
			log.Infof("download dispatcher | read task from queue: %d", taskId)

			taskIdChan <- taskId
			log.Infof("download dispatcher | send to download worker: %d", taskId)
		}
	}()

	// download result
	// 读取 resultChan
	go func() {
		for true {
			result := <-resultChan
			if result.done == false {
				log.Errorf("something wrong with task: %d", result.taskId)
				// 将 task 的状态改为 TaskDownloadError
				if result.task.Id == 0 {
					// downloader 中没有获取到 task
					err := data.TaskDownloadEnqueue(result.taskId)
					if err != nil {
						log.Errorf("download result | task error %d", result.taskId)

					}
					continue
				}
			}

			// result 正常
			log.Infof("download result | task %d done", result.taskId)
			// 设置 local file 的 status
			err := result.task.File.SetLocalFileStatus(data.FileExist)
			if err != nil {
				log.Errorf("download result | task %d set local file status error: %v", result.taskId, err)
			}
			// 设置 task 的 status
			err = result.task.SetStatus(data.TaskDownloaded)
			if err != nil {
				log.Errorf("download result | task set state to finish error %d", result.task.Id)
			}

			// 将task 推送到 mirror worker 的队列
			err = data.TaskMirrorEnqueue(result.task.Id)
			if err != nil {
				log.Errorf("download result | push to mirror worker queue error: %d", result.task.Id)
			}
		}
	}()

	// TODO: mirror dispatcher
	go func() {
		for true {
			time.Sleep(time.Second)

			// 取出 task
			taskIdStr := data.RedisCli.LPop(data.RedisMirrorTaskQueueId).Val()

			if len(taskIdStr) == 0 {
				// 当前redis队列为空
				continue
			}
			taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
			if err != nil {
				log.Errorf("mirror dispatcher | task from queue error: %s", taskIdStr)
				continue
			}
			log.Infof("mirror dispatcher | read task from queue: %d", taskId)

			mirrorTaskIdChan <- taskId
			log.Infof("mirror dispatcher | send to mirror worker: %d", taskId)
		}
	}()

	// TODO: mirror result
	go func() {
		for true {
			result := <-mirrorResultChan
			if result.done == false {
				log.Errorf("mirror result | something wrong with task: %d", result.taskId)
				// 将 task 的状态改为 TaskDownloadError
				if result.task.Id == 0 {
					// downloader 中没有获取到 task
					err := data.TaskMirrorEnqueue(result.taskId)
					if err != nil {
						log.Errorf("mirror result | task error %d", result.taskId)
					}
					continue
				}
			}

			// 设置 mirror file 的 status 为 exist
			err := result.task.File.SetMirrorFileStatus(data.FileExist)
			if err != nil {
				log.Errorf("mirror result | mirror file %d set state to exist error %v", result.task.File.Id, err)
			}

			// 设置 task 的 status 为 finished
			err = result.task.SetStatus(data.TaskFinished)
			// TODO: push task to trans to mirror queue
			if err != nil {
				log.Errorf("mirror result | task set state to finish error %d", result.task.Id)
			}
			log.Infof("mirror result | task %d done", result.taskId)

		}
	}()

	// 处理系统信号，阻塞，关闭资源并退出
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-sigs
	log.Infof("received signal: %s", sig.String())
	log.Info("close all channels...")
	close(taskIdChan)
	close(resultChan)
	log.Info("exit...")
}

// 通过header获取文件版本
// 有限通过eTag头
// 没有的话则使用当天日期对应的unix时间戳
func getVersion(resp *http.Response) string {
	var version string

	version = resp.Header.Get("eTag")
	if len(version) == 0 {
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		version = string(today.Unix())
	}
	return version
}

// 获取文件名
// 先尝试从Content-Disposition头中获取
// 如果没有则使用url的最后一段
func getFilename(resp *http.Response, u *url.URL) string {
	var filename string
	mediaInfo := resp.Header.Get("Content-Disposition")
	if len(mediaInfo) > 0 {
		// header中可能有文件名信息
		_, params, _ := mime.ParseMediaType(mediaInfo)
		filename = params["name"]
		if len(filename) > 0 {
			// header中有文件名
			return filename
		}
	}

	// header中没有文件名
	// 从url中取，这里先将url最后可能存在的 / 去掉
	pathes := strings.Split(strings.TrimSuffix(u.Path, "/"), "/")
	filename = pathes[len(pathes)-1]

	return filename
}
