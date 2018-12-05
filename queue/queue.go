package queue

import (
	commonTypes "github.com/revan730/clipper-common/types"
)

// Queue provides interface for message queue operations
type Queue interface {
	Close()
	PublishCDJob(jobMsg *commonTypes.CDJob) error
	MakeCIMsgChan() (chan []byte, error)
}
