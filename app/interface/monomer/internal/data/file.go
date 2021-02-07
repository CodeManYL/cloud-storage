package data

import (
	"github.com/cloud-storage/app/interface/monomer/internal/pkg"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/xormplus/xorm"
)

type FileData struct {
	engine *xorm.Engine
	redisPool *redis.Pool
}

func NewFileData(engine *xorm.Engine,redisPool *redis.Pool) *FileData {
	return &FileData{engine,redisPool}
}

func (fd *FileData) AddFileMetaBySha1(fileSha1,fileName,fileAddr string,fileSize int64) error {
	sql :="insert into tbl_file(file_sha1,file_name,file_size,file_addr) values (?,?,?,?)"
	_, err := fd.engine.Exec(sql, fileSha1,fileName,fileSize,fileAddr)
	return err
}

func (fd *FileData) AddUserFile(username,sha1,fileName string,size int64) error {
	sql := "insert into tbl_user_file(user_name,file_sha1,file_name,file_size) values(?,?,?,?)"
	_, err := fd.engine.Exec(sql,username,sha1,fileName,size)
	return err
}

func (fd *FileData) GetFileMetaBySha1(fileSha1 string) (fileMeta *FileMeta, err error){
	fileMeta = &FileMeta{}
	ok, err := fd.engine.SQL("select file_sha1,file_name,file_size,file_addr,update_at,create_at from tbl_file where file_sha1=?", fileSha1).Get(fileMeta)
	if err != nil {
		return nil,err
	}
	if !ok {
		err = pkg.ErrNotFound
	}
	return
}

func (fd *FileData) UpdateFileMetaBySha1(fileSha1,fileName string) error {
	sql := "update tbl_file set file_name = ? where file_sha1 = ?"
	_, err := fd.engine.Exec(sql,fileName,fileSha1)
	return err
}


