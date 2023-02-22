package seed

import (
	"torrentClient/models"
)

type SeedRequest struct {
	Peer    *models.Peer
	Message *models.Message
}
