package common

import (
	"fmt"
	"io"
	"os"
	"torrentClient/models"

	"github.com/IncSW/go-bencode"
)

func LoadOrCreateDownloadBlob(manifest *models.Manifest) *os.File {
	os.MkdirAll(manifest.Name, 0700)
	blobFilePath := manifest.Name + "/" + manifest.Name + ".blob"

	if _, err := os.Stat(blobFilePath); os.IsNotExist(err) {
		_, err = os.Create(blobFilePath)
		if err != nil {
			panic(err)
		}
	}

	blobFile, err := os.OpenFile(blobFilePath, os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Can't create file", manifest.Name, err)
		panic(err)
	}

	// err = blobFile.Truncate(manifest.Length)
	if err != nil {
		fmt.Println("Can't truncate file", manifest.Name, err)
		panic(err)
	}

	return blobFile
}

func WriteBlobToFiles(manifest *models.Manifest) {
	blobFilePath := manifest.Name + "/" + manifest.Name + ".blob"

	if _, err := os.Stat(blobFilePath); os.IsNotExist(err) {
		_, err = os.Create(blobFilePath)
		if err != nil {
			panic(err)
		}
	}

	blobFile, err := os.OpenFile(blobFilePath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Can't open blob file", manifest.Name, err)
		panic(err)
	}

	defer blobFile.Close()

	for _, file := range manifest.FileInfos {
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			_, err = os.Create(file.Path)
			if err != nil {
				panic(err)
			}
		}

		f, err := os.OpenFile(file.Path, os.O_RDWR, 0644)

		if err != nil {
			fmt.Println("Can't create file", file.Path, err)
			panic(err)
		}

		err = f.Truncate(file.Length)
		if err != nil {
			fmt.Println("Can't truncate file", file.Path, err)
			panic(err)
		}
		blobFile.Seek(file.Offset, 0)

		_, err = io.CopyN(f, blobFile, file.Length)

		if err != nil {
			fmt.Println("Can't copy file", file.Path, err)
			panic(err)
		}
		f.Close()
	}
}

func ReadManifestFromFile(filePath string) models.Manifest {
	content, err := os.ReadFile(filePath)
	if err != nil {
		println("Can't open torrent file", err)
		panic(err)
	}
	data, err := bencode.Unmarshal(content)
	if err != nil {
		panic(err)
	}
	return models.DecodeManifestFile(data)
}
