package biz

import (
	"bufio"
	"fmt"
	"github.com/cloud-storage/app/interface/monomer/internal/data"
	"mime/multipart"
	"os"
)

type FileBiz struct {
	fileData *data.FileData
}

func NewFileBiz(fileData *data.FileData) *FileBiz{
	return &FileBiz{fileData}
}

//AddFileMetaAndUserFileInfo  存储文件元信息和用户文件信息到db
func (fb *FileBiz) AddFileMetaAndUserFileInfo(sha1,fileName,path,username string,size int64) (err error) {
	if err = fb.fileData.AddFileMetaBySha1(sha1,fileName,path,size);err != nil {
		fmt.Println(err)
		return
	}

	if err = fb.fileData.AddUserFile(username,sha1,fileName,size);err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (fb *FileBiz) GetFileMetaBySha1(fileSha1 string) (*data.FileMeta,error) {
	return fb.fileData.GetFileMetaBySha1(fileSha1)
}

func (fb *FileBiz) UpdateFileMetaBySha1(fileSha1,fileName string) error {
	return fb.fileData.UpdateFileMetaBySha1(fileSha1,fileName)
}

func (fb *FileBiz) AddUserFile(username,sha1,fileName string,size int64) error {
	return fb.fileData.AddUserFile(username,sha1,fileName,size)
}


func (fb *FileBiz) AddMultipartUploadInfoCache(m *data.MultipartUploadInfo) error{
	return fb.fileData.AddMultipartUploadInfo(m)
}

func (fb *FileBiz) Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (fb *FileBiz) UploadPart(uploadId,chunkIndex string,f multipart.File) (err error) {
	dir := "./app/interface/data/" + uploadId + "/"
	ok := fb.Exists(dir)
	if !ok {
		if err = os.MkdirAll(dir, 0755); err != nil {
			fmt.Println(err)
			return
		}
	}

	path := dir + chunkIndex
	f2, err := os.Create(path)
	defer f2.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 1024*1024)
	reader := bufio.NewReader(f)
	writer := bufio.NewWriterSize(f2,5)
	defer writer.Flush()

	for {
		n, err := reader.Read(buf)
		writer.Write(buf[:n])
		if err != nil {
			break
		}
	}

	//写cache
	if err = fb.fileData.AddMultipartUploadChunkIndex(uploadId,chunkIndex);err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (fb *FileBiz) MergeFile(upId,userName,fileHash,fileName string,fileSize int64) (ok bool,err error){
	ok,err = fb.fileData.GetMultipartUploadFileInfoByUpload(upId)
	if err != nil  {
		return
	}
	if !ok {
		return
	}

	//更新文件表和用户表
	if err = fb.fileData.AddUserFile(userName,fileHash,fileName,fileSize);err != nil {
		fmt.Println(err)
		return
	}

	if err = fb.fileData.AddFileMetaBySha1(fileHash,fileName,"",fileSize);err != nil {
		fmt.Println(err)
		return
	}

	return
}

