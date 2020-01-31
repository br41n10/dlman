package data

import (
	"database/sql"
	"dlman/config"
	_ "dlman/config"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"net/url"
	"os"
	"path"
)

type fileStatus string
type Carrier string

const (
	FileCreated  fileStatus = "CREATED"
	FileExist    fileStatus = "EXIST"
	FileNotExist fileStatus = "NOT_EXIST"
	FileDeleted  fileStatus = "DELETED"
)

const (
	CarrierLocal Carrier = "LOCAL"
)

type File struct {
	Id           int64
	uuid         uuid.UUID
	Name         sql.NullString // meta data
	Version      sql.NullString
	DownloadTime int64
	local        localFile
	Mirror       MirrorFile
}

// 本地文件
type localFile struct {
	file   *os.File
	status fileStatus
}

// 镜像文件
type MirrorFile struct {
	Url     *url.URL
	Status  fileStatus
	Carrier Carrier
}

// 通过localFile生成文件地址
// 这里的前提是我们的文件是按照uuid放在LocalVendorPath下的
func (f File) LocalAbsPath() string {
	// TODO: 如果后面再localFile中不再保存os.file的话，这里修改成localFile的方法
	return path.Join(config.LocalVendorPath, f.uuid.String())
}

func LocalAbsPath(baseDir string, u uuid.UUID) string {
	return path.Join(baseDir, u.String())
}

func NewFile() (*File, error) {

	// TODO: 如果后面localFile中不再保存os.file的话，把file结构体的创建提到最前面

	// 根据随机的uuid生成文件名
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	// 创建文件
	filePath := LocalAbsPath(config.LocalVendorPath, u)
	fmt.Printf("待创建本地文件路径：%s\n", filePath)

	// 创建文件
	osFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	// 这里不需要写文件，则关闭掉
	err = osFile.Close()
	if err != nil {
		fmt.Println(err)
	}

	// 生成localFile结构体
	localFile := localFile{
		file:   osFile,
		status: FileCreated,
	}

	// 生成 File 结构体
	f := File{
		uuid:  u,
		local: localFile,
	}

	// 写到数据库中
	result, err := Db.Exec(
		"insert into File (uuid, mirror_file_status, local_file_status) values (?, ?, ?)",
		u.String(), FileNotExist, FileCreated)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	f.Id = id

	return &f, nil
}

func GetFileById(id int64) (*File, error) {
	file := &File{}

	// 中间变量
	var (
		rawMirrorUrl sql.NullString
		rawUuid      string
	)

	row := Db.QueryRow("select id, name, uuid, version, mirror_url, mirror_file_status, local_file_status, download_time, mirror_carrier from File where id = ?", id)
	err := row.Scan(&file.Id, &file.Name, &rawUuid, &file.Version, &rawMirrorUrl, &file.Mirror.Status, &file.local.status, &file.DownloadTime, &file.Mirror.Carrier)
	if err != nil {
		return nil, err
	}

	file.uuid, _ = uuid.Parse(rawUuid) // 这里不会错，因为是正确的uuid String() 进去的

	if rawMirrorUrl.Valid {
		file.Mirror.Url, _ = url.Parse(rawMirrorUrl.String)
	} else {
		file.Mirror.Url = &url.URL{}
	}

	return file, nil
}

func (f *File) SetMirrorUrl(u *url.URL) error {
	f.Mirror.Url = u
	f.Mirror.Status = FileExist
	_, err := Db.Exec("update File set mirror_url = ?, mirror_file_status = ? where id = ?", f.Mirror.Url.String(), f.Mirror.Status, f.Id)
	if err != nil {
		log.Errorf("File|SetMirrorUrl|id: %d, url: %s, status: %s", f.Id, f.Mirror.Url.String(), f.Mirror.Status)
		return err
	}
	return nil
}

func (f *File) SetCarrier(c Carrier) error {
	f.Mirror.Carrier = c
	_, err := Db.Exec("update File set mirror_carrier = ? where id = ?", c, f.Id)
	if err != nil {
		log.Errorf("File|SetCarrier|id: %d, carrier: %s", f.Id, c)
		return err
	}
	return nil
}

func (f *File) SetLocalFileStatus(fs fileStatus) error {
	_, err := Db.Exec("update File set local_file_status = ? where id = ?", FileExist, f.Id)
	if err != nil {
		log.Errorf("File|SetLocalFileStatus|id: %d, file statue: %s", f.Id, fs)
		return err
	}
	return nil
}

func (f *File) SetMirrorFileStatus(fs fileStatus) error {
	_, err := Db.Exec("update File set mirror_file_status = ? where id = ?", FileExist, f.Id)
	if err != nil {
		log.Errorf("File|SetMirrorFileStatus|id: %d, file statue: %s", f.Id, fs)
		return err
	}
	return nil
}

func (c *Carrier) Scan(value interface{}) error {
	if value == nil {
		// 如果数据库中是 null ，则什么都不用做，因为该字段已经被初始化成 "" 了
		return nil
	}
	v := fmt.Sprintf("%v", value)
	*c = Carrier(v)
	return nil
}
