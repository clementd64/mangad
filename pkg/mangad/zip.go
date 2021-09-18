package mangad

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"time"
)

func zipFolder(folder, filename string) error {
	dir, err := os.ReadDir(folder)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for _, file := range dir {
		fileToZip, err := os.Open(path.Join(folder, file.Name()))
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {
			return err
		}

		writer, err := zw.CreateHeader(&zip.FileHeader{
			Name:               info.Name(),
			UncompressedSize64: uint64(info.Size()),
			Modified:           time.UnixMilli(0),
		})
		if err != nil {
			return err
		}

		if _, err = io.Copy(writer, fileToZip); err != nil {
			return err
		}
	}

	return zw.Close()
}
