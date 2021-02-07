package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-storage/app/interface/monomer/internal/biz"
	"github.com/cloud-storage/app/interface/monomer/internal/data"
	"github.com/cloud-storage/app/interface/monomer/internal/pkg"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type FileService struct {
	fileBiz *biz.FileBiz
}

type ResData struct {
	Code int
	Msg  string
	Data interface{}
}

func NewFileService(fileBiz *biz.FileBiz) *FileService {
	return &FileService{fileBiz}
}

func (fs *FileService) Upload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	f, h, err := request.FormFile("file")
	defer f.Close()

	if err != nil {
		writer.Write([]byte("失败"))
		return
	}

	if err := request.ParseMultipartForm(32 >> 20); err != nil {
		fmt.Println(err)
		return
	}

	username := request.Form.Get("username")
	fileName := h.Filename
	path := "./app/interface/" + fileName
	f2, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	defer f2.Close()

	if err != nil {
		writer.Write([]byte("失败"))
		return
	}

	size, err := io.Copy(f2, f)

	if err != nil {
		fmt.Println(err)
		return
	}

	f2.Seek(0, 0)
	sha1 := pkg.FileSha1(f2)

	if err = fs.fileBiz.AddFileMetaAndUserFileInfo(sha1, fileName, path, username, size); err != nil {
		return
	}
	writer.Write([]byte("ok"))
}

func (fs *FileService) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}

	fileSha1List := r.Form["fileSha1"]
	if fileSha1List == nil {
		fmt.Println("参数不正确")
		return
	}

	fileSha1 := fileSha1List[0]
	fileMeta, err := fs.fileBiz.GetFileMetaBySha1(fileSha1)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := json.Marshal(fileMeta)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Write(res)
}

func (fs *FileService) Download(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println("参数解析错误")
		return
	}

	fileSha1 := r.Form.Get("fileSha1")
	fileMeta, err := fs.fileBiz.GetFileMetaBySha1(fileSha1)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Open(fileMeta.Location)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-type", "application/octect-stream")
	w.Header().Set("Content-disposition", "attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(body)
}

func (fs *FileService) UpdateFileMetaBySha1(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}

	fileName := r.Form.Get("fileName")
	fileHash := r.Form.Get("fileHash")

	if err := fs.fileBiz.UpdateFileMetaBySha1(fileHash, fileName); err != nil {
		w.Write([]byte("更新失败"))
		return
	}

	w.Write([]byte("操作成功"))
}

func (fs *FileService) DeleteFileMetaBySha1(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}

	fileHash := r.Form.Get("fileHash")
	fileMeta, err := fs.fileBiz.GetFileMetaBySha1(fileHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	if fileMeta == nil {
		w.Write([]byte("不存在"))
		return
	}

	if err := os.Remove(fileMeta.Location); err != nil {
		fmt.Println(err)
		return
	}

	w.Write([]byte("Ok"))
}

func (fs *FileService) FastUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	sha1 := r.Form.Get("hash")
	username := "admin"

	//触发秒传
	fileMeta, err := fs.fileBiz.GetFileMetaBySha1(sha1)
	ok := errors.Is(err, pkg.ErrNotFound)
	if err != nil && !ok {
		fmt.Println(err)
		return
	}

	if ok {
		//触发失败
		res, err := json.Marshal(&ResData{
			404,
			"触发失败",
			nil,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Write(res)
		return
	}

	if err := fs.fileBiz.AddUserFile(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize); err != nil {
		fmt.Println(err)
		return
	}

	res := &ResData{
		Code: 200,
		Msg:  "OK",
	}

	resByte, _ := json.Marshal(res)

	w.Write(resByte)
}

func (fs *FileService) InitMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	r.ParseForm()
	//参数
	fileHash := r.Form.Get("fileHash")
	fileSize, err := strconv.Atoi(r.Form.Get("fileSize"))
	if err != nil {
		w.Write([]byte("参数错误"))
		return
	}
	username := r.Form.Get("username")

	m := &data.MultipartUploadInfo{
		FileHash:   fileHash,
		FileSize:   fileSize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
	}
	if err := fs.fileBiz.AddMultipartUploadInfoCache(m); err != nil {
		fmt.Println(err)
		return
	}

	w.Write([]byte("ok"))
}

// UploadPart 分块上传接口
func (fs *FileService) UploadPartHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(32 >> 20); err != nil {
		fmt.Println(err)
		return
	}

	//username := "admin"
	uploadId := r.Form.Get("uploadId")
	chunkIndex := r.Form.Get("chunkIndex")
	f, _, err := r.FormFile("file")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = fs.fileBiz.UploadPart(uploadId,chunkIndex,f);err != nil {
		fmt.Println(err)
		return
	}

	w.Write([]byte("ok"))
}

func (fs *FileService) PostFromParse(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(32 >> 20); err != nil {
		fmt.Println(err)
		return
	}
}


// MergeFileHandle  合并分块上传的文件
func (fs *FileService) MergeFileHandle(w http.ResponseWriter, r *http.Request) {
	fs.PostFromParse(w,r)
	upId := r.Form.Get("upId")
	fileHash := r.Form.Get("fileHash")
	fileSize, err := strconv.Atoi(r.Form.Get("fileSize"))
	if err != nil {
		w.Write([]byte("参数错误"))
		return
	}
	fileName := r.Form.Get("fileName")
	userName := r.Form.Get("userName")

	ok,err := fs.fileBiz.MergeFile(upId,userName,fileHash,fileName,int64(fileSize))
	if err != nil {
		fmt.Println(err)
		return
	}
	if !ok {
		w.Write([]byte("not ok"))
		return
	}

	w.Write([]byte("ok"))
}