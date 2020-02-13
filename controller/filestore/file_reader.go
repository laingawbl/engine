package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/laingawbl/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

var openFileReader = fsOpenFileReader

type reader interface {
	ReadBytes(delim byte) ([]byte, error)
	Close() error
}

type fsReader struct {
	*os.File
	*bufio.Reader
}

func fsOpenFileReader(dir string, id string) (reader, error) {
	f, err := os.OpenFile(getFilePath(dir, id), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &fsReader{
		File:   f,
		Reader: bufio.NewReader(f),
	}, nil
}

func readLine(r reader, out interface{}) (bool, error) {
	bytes, err := r.ReadBytes('\n')
	eof := err == io.EOF

	if err != nil && !eof {
		return false, err
	}

	if !containsJSONObject(string(bytes)) {
		return !eof, errors.New("invalid json")
	}

	if err = json.Unmarshal(bytes, out); err != nil {
		return !eof, err
	}

	return !eof, nil
}

func readGameInfo(r reader) (*pb.Game, bool, error) {
	info := &pb.Game{}
	more, err := readLine(r, info)
	return info, more, err
}

func readFrame(r reader) (*pb.GameFrame, bool) {
	for {
		f := &pb.GameFrame{}
		more, err := readLine(r, f)

		// It worked so return result
		if err == nil {
			return f, more
		}

		// This line wasn't a frame and reached end of file
		if !more {
			return nil, false
		}
	}
}

func readArchiveHeader(dir string, id string) (*pb.Game, error) {
	r, err := openFileReader(dir, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = r.Close()
		if err != nil {
			log.WithError(err).Error("Error while closing file reader")
		}
	}()

	info, _, err := readGameInfo(r)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func containsJSONObject(s string) bool {
	trimmed := strings.TrimSpace(s)
	return len(trimmed) > 1 && trimmed[0] == '{'
}

func readArchive(dir string, id string) (gameArchive, error) {
	r, err := openFileReader(dir, id)
	if err != nil {
		return gameArchive{}, err
	}
	defer func() {
		err = r.Close()
		if err != nil {
			log.WithError(err).Error("Error while closing reader")
		}
	}()

	// Skip over game info line
	game, moreLines, err := readGameInfo(r)
	if err != nil {
		return gameArchive{}, err
	}

	// Read the actual frames
	frames := []*pb.GameFrame{}
	for moreLines {
		f, more := readFrame(r)
		moreLines = more
		if f != nil {
			frames = append(frames, f)
		}
	}

	return gameArchive{
		game:   game,
		frames: frames,
	}, nil
}

// ReadGameFrames loads all the game frames stored in given file.
func ReadGameFrames(dir string, id string) ([]*pb.GameFrame, error) {
	archive, err := readArchive(dir, id)
	if err != nil {
		return nil, err
	}

	return archive.frames, nil
}

// ReadGameInfo reads the header info from the given file.
func ReadGameInfo(dir string, id string) (*pb.Game, error) {
	return readArchiveHeader(dir, id)
}
