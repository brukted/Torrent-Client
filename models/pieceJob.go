package models

type PieceJob struct {
	PieceIndex  int
	PieceHash   [20]byte
	PieceLength int
}

type PieceJobResult struct {
	PieceIndex int
	PieceData  []byte
}

type PieceJobProgress struct {
	PieceIndex      int
	Buffer          []byte
	TotalDownloaded int
	PieceLength     int
}
