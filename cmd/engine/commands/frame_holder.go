package commands

import (
	"sync"

	"github.com/laingawbl/engine/controller/pb"
)

type frameHolder struct {
	sync.RWMutex
	frames []*pb.GameFrame
	ffc    chan *pb.GameFrame
}

func (fh *frameHolder) append(frame *pb.GameFrame) {
	fh.Lock()
	defer fh.Unlock()

	if len(fh.frames) == 0 {
		if fh.ffc == nil {
			fh.ffc = make(chan *pb.GameFrame)
		}
		fh.ffc <- frame
		close(fh.ffc)
	}

	fh.frames = append(fh.frames, frame)
}

func (fh *frameHolder) get(index int) *pb.GameFrame {
	fh.RLock()
	defer fh.RUnlock()

	if index < 0 || index >= len(fh.frames) {
		return nil
	}

	return fh.frames[index]
}

func (fh *frameHolder) initialFrame() <-chan *pb.GameFrame {
	if fh.ffc == nil {
		fh.ffc = make(chan *pb.GameFrame)
	}
	return fh.ffc
}

func (fh *frameHolder) count() int {
	fh.RLock()
	defer fh.RUnlock()

	return len(fh.frames)
}
