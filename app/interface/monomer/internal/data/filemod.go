package data

type  FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string  `xorm:"varchar(1024) 'file_addr'"`
	UploadAt string
}

type UserFile struct {
	UserName string
	FileSha1 string
	FileSize int64
	FileName string
	UploadAt string
	LastUpdate string
}

// MultipartUploadInfo 缓存文件分块的元信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string          //上传的唯一Id
	ChunkSize int            //每一块的尺寸
	ChunkCount int           //有多少块
}