package task

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"path/filepath"

	"github.com/disintegration/imaging"
	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/storage"
)

type ThumbnailTask struct {
	fileDAO *dao.FileDAO
	storage storage.Interface
	fileID  int
}

func NewThumbnailTask(fileDAO *dao.FileDAO, storage storage.Interface, fileID int) *ThumbnailTask {
	return &ThumbnailTask{
		fileDAO: fileDAO,
		storage: storage,
		fileID:  fileID,
	}
}

func (t *ThumbnailTask) Execute() error {
	log.Printf("[TASK] Generating thumbnail for file %d", t.fileID)

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

	src, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	dst := imaging.Resize(src, 400, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, dst, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return err
	}

	thumbPath := filepath.Join(filepath.Dir(file.StoragePath), "thumb_"+filepath.Base(file.StoragePath))
	if err := t.storage.Save(thumbPath, &buf); err != nil {
		return err
	}

	return t.fileDAO.UpdateThumbPath(t.fileID, thumbPath)
}
