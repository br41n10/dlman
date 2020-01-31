package data

import (
	"net/url"
	"testing"
)

func TestGetTaskById(t *testing.T) {
	task, err := GetTaskById(6)
	if err != nil {
		t.Errorf("Test|GetTaskById|FAIL, err: %v", err)
	}
	t.Logf("Task: %v", task)
}

func TestNewTask(t *testing.T) {
	url, _ := url.Parse("https://cdn.bootcss.com/jquery/3.4.1/jquery.js")
	task, err := NewTask(url)
	if err != nil {
		t.Error(err)
	}
	t.Logf("new Task id: %d", task.Id)
}

func TestTask_Start(t *testing.T) {
	task, err := GetTaskById(6)
	if err != nil {
		t.Error(err)
	}
	err = task.Start()
	if err != nil {
		t.Error(err)
	}
}
