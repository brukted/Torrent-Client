package models

import (
	"crypto/sha1"
	"path"

	"github.com/IncSW/go-bencode"
)

type Manifest struct {
	PieceHashes [][20]byte
	// Main tracker
	Announce string
	// Backup trackers
	AnnounceList []string
	InfoHash     [20]byte
	PieceLength  int64
	Length       int64
	Name         string
	Comment      string
	CreatedBy    string
	FileInfos    []FileInfo
}

type FileInfo struct {
	Path   string
	Name   string
	Length int64
	Offset int64
}

func DecodeManifestFile(data interface{}) Manifest {
	manifestoMap := data.(map[string]interface{})

	announce := string(manifestoMap["announce"].([]byte))

	announceList := []string{}
	if manifestoMap["announce-list"] != nil {
		for _, announce := range manifestoMap["announce-list"].([]interface{}) {
			announceList = append(announceList, string(announce.([]interface{})[0].([]byte)))
		}
	}

	comment := ""
	if manifestoMap["comment"] != nil {
		comment = string(manifestoMap["comment"].([]byte))
	}

	createdBy := ""
	if manifestoMap["created by"] != nil {
		createdBy = string(manifestoMap["created by"].([]byte))
	}

	info := manifestoMap["info"].(map[string]interface{})

	pieceLength := info["piece length"].(int64)
	name := string(info["name"].([]byte))

	// NOTICE: Not sure if this is correct
	infoBytes, _ := bencode.Marshal(info)
	infoHash := sha1.Sum(infoBytes)

	pieceHashes := [][20]byte{}
	pieces := info["pieces"].([]byte)

	for i := 0; i < len(pieces); i += 20 {
		var currentHash [20]byte
		copy(currentHash[:], pieces[i:i+20])
		pieceHashes = append(pieceHashes, currentHash)
	}

	files := []interface{}{info}

	if info["files"] != nil {
		files = info["files"].([]interface{})
	}

	fileInfos := []FileInfo{}
	var offset int64

	for _, file := range files {
		file := file.(map[string]interface{})

		parts := []string{name}

		if file["path"] != nil {
			for _, part := range file["path"].([]interface{}) {
				parts = append(parts, string(part.([]byte)))
			}
		} else {
			parts = append(parts, name)
		}

		length := file["length"].(int64)

		fileInfos = append(fileInfos, FileInfo{
			Path:   path.Join(parts...),
			Name:   parts[len(parts)-1],
			Length: length,
			Offset: offset,
		})

		offset += length
	}

	return Manifest{
		Announce:     announce,
		AnnounceList: announceList,
		InfoHash:     infoHash,
		PieceHashes:  pieceHashes,
		PieceLength:  pieceLength,
		Length:       offset,
		Name:         name,
		Comment:      comment,
		CreatedBy:    createdBy,
		FileInfos:    fileInfos,
	}
}
