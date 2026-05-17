package task

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"image"
	"io"
	"log"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/phash"
	"qalbum-server/pkg/storage"
)

type PHashTask struct {
	fileDAO *dao.FileDAO
	storage storage.Interface
	fileID  int
}

func NewPHashTask(fileDAO *dao.FileDAO, storage storage.Interface, fileID int) *PHashTask {
	return &PHashTask{
		fileDAO: fileDAO,
		storage: storage,
		fileID:  fileID,
	}
}

func (t *PHashTask) Execute() error {
	log.Printf("[TASK] Computing pHash for file %d", t.fileID)

	file, err := t.fileDAO.GetByID(t.fileID)
	if err != nil {
		return err
	}

	if file.FileType != "image" {
		return nil
	}

	reader, err := t.storage.Open(file.StoragePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return err
	}

	pHash := phash.ComputeDCT(img)
	pHashHex := formatPHash(pHash)

	return t.fileDAO.UpdatePHash(t.fileID, pHashHex)
}

func formatPHash(hash uint64) string {
	return fmt.Sprintf("%016x", hash)
}
