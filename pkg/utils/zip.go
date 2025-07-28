package utils

import (
	"archive/zip"
	"io"
	"os"
)

func AddFileToZip(zipWriter *zip.Writer, filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
