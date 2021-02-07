package data

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
)

// AddMultipartUploadInfo 添加分块上传的初始化文件信息
func (fd *FileData) AddMultipartUploadInfo(m *MultipartUploadInfo) (err error){
	c := fd.redisPool.Get()
	defer c.Close()

	_,err = c.Do("HMSET","MP_"+m.UploadID,"chunkcount",m.ChunkCount,"filehash",m.FileHash,"filesize",m.FileSize)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

// AddMultipartUploadInfo 添加分块上传的初始化文件信息
func (fd *FileData) AddMultipartUploadChunkIndex(uploadId,chkIndex string) (err error){
	c := fd.redisPool.Get()
	defer c.Close()

	_,err = c.Do("HMSET",uploadId,"chkIndex_"+chkIndex,1)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

//GetMultipartUploadFileInfoByUpload  获得upid所有字段信息
func (fd *FileData) GetMultipartUploadFileInfoByUpload(upId string) (ok bool,err error) {
	c := fd.redisPool.Get()
	defer c.Close()

	all,err :=redis.Values(c.Do("HGETALL",upId))
	if err != nil {
		fmt.Println(err)
		return
	}

	num := 0
	cur := 0

	for i := 0; i < len(all); i += 2 {
		key := string(all[i].([]byte))
		value := string(all[i+1].([]byte))

		if key == "chunkcount" {
			num,_ = strconv.Atoi(value)
		} else if strings.HasPrefix(key,"chkIndex_") && value == "1" {
			n,_ := strconv.Atoi(value)
			cur += n
		}
	}

	if num == cur {
		ok = true
		return
	}

	return


}

