package main

import (
	"fmt"
	"github.com/cloud-storage/app/interface/monomer/internal/biz"
	"github.com/cloud-storage/app/interface/monomer/internal/data"
	"github.com/cloud-storage/app/interface/monomer/internal/service"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
)




func HTTPInterceptor(f func(w http.ResponseWriter, r *http.Request))func(w http.ResponseWriter, r *http.Request){
	return  func(w http.ResponseWriter, r *http.Request){
		//闭包不但是可以用外部函数里面的变量
		//执行外面的时候手动把里面执行了
		//这里的闭包没有外部变量，而是伪装了一个对象，类似DNS劫持
		fmt.Println("拦截")
		f(w,r)
	}
}

func TestHandle(w http.ResponseWriter, r *http.Request){
	fmt.Println(1)
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 20,
		MaxActive:30,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func () (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}


func main() {

	engine, err := xorm.NewEngine("mysql", "root:33css@/cloud-storage?charset=utf8")
	if err != nil {
		panic(err)
	}

	redisPool := newPool()
	defer redisPool.Close()

	fileData1 := data.NewFileData(engine,redisPool)
	fileBiz := biz.NewFileBiz(fileData1)
	fileServer := service.NewFileService(fileBiz)

	http.HandleFunc("/upload",fileServer.Upload)

	http.HandleFunc("/get-file-info", fileServer.GetFileInfo)

	http.HandleFunc("/download", fileServer.Download)

	http.HandleFunc("/update", fileServer.UpdateFileMetaBySha1)

	http.HandleFunc("/delete", fileServer.DeleteFileMetaBySha1)

	http.HandleFunc("/fast-upload", fileServer.FastUpload)

	http.HandleFunc("/test",HTTPInterceptor(TestHandle))

	http.HandleFunc("/init-multipart-upload",fileServer.InitMultipartUpload)

	http.HandleFunc("/upload-part",fileServer.UploadPartHandle)

	http.HandleFunc("/merge",fileServer.MergeFileHandle)

	log.Fatal(http.ListenAndServe(":8000",nil))

}
