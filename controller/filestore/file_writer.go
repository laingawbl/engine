package filestore

import (
	"encoding/json"
	"os"

	"github.com/laingawbl/engine/controller/pb"
)

var openFileWriter = appendOnlyFileWriter

type writer interface {
	WriteString(s string) (int, error)
	Close() error
}

func requireSaveDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}

func writeLine(w writer, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.WriteString(string(j) + "\n")
	return err
}

func writeFrame(w writer, f *pb.GameFrame) error {
	return writeLine(w, f)
}

func writeGameInfo(w writer, game *pb.Game, snakes []*pb.Snake) error {
	return writeLine(w, game)
}

func appendOnlyFileWriter(dir string, id string, mustCreate bool) (writer, error) {
	if err := requireSaveDir(dir); err != nil {
		return nil, err
	}

	path := getFilePath(dir, id)
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	if mustCreate {
		flags |= os.O_EXCL
	}
	return os.OpenFile(path, flags, 0600)
}
