package ipfs

type IpfsUploader interface {
	UploadBytes(data []byte) (string, error)
}

func UploadBytes(uploader IpfsUploader, data []byte) (string, error) {
	return uploader.UploadBytes(data)
}

// TODO: Currently providing filebase support to upload on ipfs in future more integration will be added
