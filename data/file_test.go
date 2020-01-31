package data

import (
	"fmt"
	"testing"
)

func TestNewFile(t *testing.T) {
	file, err := NewFile()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(file.Id, file.Name, file.Version, file.LocalAbsPath(), file.Mirror.Url, file.Mirror.Status, file.local.status, file.DownloadTime)
}

func TestGetFileById(t *testing.T) {
	t.Logf("query File by id 11")
	file, err := GetFileById(13)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(file.Id, file.Name, file.Version, file.LocalAbsPath(), file.Mirror.Url, file.Mirror.Status, file.local.status, file.DownloadTime)
}

func TestLocalAbsPath(t *testing.T) {
	file, err := GetFileById(26)
	if err != nil {
		t.Error(err)
	}
	path := file.LocalAbsPath()
	if len(path) == 0 {
		t.Fail()
	}
	t.Logf("path of file %s is %s", file.uuid.String(), path)
}
