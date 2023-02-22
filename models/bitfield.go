package models

import (
	"math"
	"os"
)

type Bitfield []byte

func (bitfield Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8

	return bitfield[byteIndex]>>(7-offset)&1 != 0
}

func (bitfield Bitfield) MarkPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	bitfield[byteIndex] |= 1 << (7 - offset)
}

func (bitfield *Bitfield) WriteToFile(manifest *Manifest, bitfieldFile *os.File) {
	bitfieldFile.WriteAt(*bitfield, 0)
}

func LoadOrCreateBitFieldFromFile(manifest *Manifest) (*Bitfield, *os.File) {
	bitfield := make(Bitfield, int(math.Ceil(float64(manifest.PieceLength)/8.0)))
	bitfieldFilePath := manifest.Name + ".bitfield"

	if _, err := os.Stat(bitfieldFilePath); os.IsNotExist(err) {
		bitfieldFile, err := os.Create(bitfieldFilePath)
		if err != nil {
			panic(err)
		}
		return &bitfield, bitfieldFile
	}

	bitfieldFile, err := os.OpenFile(bitfieldFilePath, os.O_RDWR, 0644)

	if err != nil {
		panic(err)
	}

	bitfieldFile.Truncate(int64(len(bitfield)))

	_, err = bitfieldFile.ReadAt(bitfield, 0)

	if err != nil {
		panic(err)
	}

	return &bitfield, bitfieldFile

}
