package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

type mockReader struct {
	*bufio.Reader
}

func (m *mockReader) Close() error {
	return nil
}

func newMockReader(text string) *mockReader {
	return &mockReader{
		Reader: bufio.NewReader(strings.NewReader(text)),
	}
}

type failReader struct{}

func (f *failReader) ReadBytes(delimiter byte) ([]byte, error) {
	return nil, errors.New("FAIL")
}

func (f *failReader) Close() error {
	return errors.New("FAIL")
}

func fileOpener(files map[string]string) func(string, string) (reader, error) {
	return func(dir string, id string) (reader, error) {
		text, ok := files[id]
		if !ok {
			return nil, errors.New("file not found")
		}
		return newMockReader(text), nil
	}
}

func gameInfoTestJSON() string {
	info := &pb.Game{
		ID:     "myid",
		Width:  10,
		Height: 12,
	}

	infoJSON, _ := json.Marshal(info)
	return string(infoJSON) + "\n"
}

func framesTestJSON() string {
	f1 := &pb.GameFrame{
		Turn: 1,
		Snakes: []*pb.Snake{
			&pb.Snake{
				ID:     "snake1",
				Body:   []*pb.Point{&pb.Point{X: 1, Y: 1}, &pb.Point{X: 1, Y: 2}},
				Health: 100,
				Death:  nil,
			},
			&pb.Snake{
				ID:     "snake2",
				Body:   []*pb.Point{&pb.Point{X: 3, Y: 1}, &pb.Point{X: 3, Y: 2}},
				Health: 100,
				Death:  nil,
			},
		},
		Food: []*pb.Point{&pb.Point{X: 6, Y: 6}},
	}

	f2 := &pb.GameFrame{
		Turn: 2,
		Snakes: []*pb.Snake{
			&pb.Snake{
				ID:     "snake1",
				Body:   []*pb.Point{&pb.Point{X: 1, Y: 2}, &pb.Point{X: 1, Y: 3}},
				Health: 100,
				Death:  nil,
			},
			&pb.Snake{
				ID:     "snake2",
				Body:   []*pb.Point{&pb.Point{X: 3, Y: 2}, &pb.Point{X: 3, Y: 3}},
				Health: 100,
				Death:  &pb.Death{Cause: "test", Turn: 2},
			},
		},
		Food: []*pb.Point{&pb.Point{X: 6, Y: 6}},
	}

	j1, _ := json.Marshal(f1)
	j2, _ := json.Marshal(f2)

	return string(j1) + "\n" + string(j2) + "\n"
}

func TestReadGameFramesBadReader(t *testing.T) {
	openFileReader = func(dir string, id string) (reader, error) {
		return &failReader{}, nil
	}
	_, err := ReadGameFrames("", "myid")

	require.NotNil(t, err)
}

func TestReadGameFramesOpenReaderError(t *testing.T) {
	openFileReader = func(dir string, id string) (reader, error) {
		return nil, errors.New("fail")
	}
	_, err := ReadGameFrames("", "myid")

	require.NotNil(t, err)
}

func TestReadGameFramesWithoutHeader(t *testing.T) {
	j := framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 1, "first frame is in header spot and should be ignored")
}

func TestReadGameFrames(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 2)
	require.Equal(t, int32(1), frames[0].Turn)
	require.Equal(t, int32(2), frames[1].Turn)
	require.Equal(t, "snake1", frames[0].Snakes[0].ID)
	require.Equal(t, "snake2", frames[0].Snakes[1].ID)
	require.Equal(t, "snake1", frames[1].Snakes[0].ID)
	require.Equal(t, "snake2", frames[1].Snakes[1].ID)
	require.Nil(t, frames[0].Snakes[0].Death)
	require.Nil(t, frames[0].Snakes[1].Death)
	require.Nil(t, frames[1].Snakes[0].Death)
	require.NotNil(t, frames[1].Snakes[1].Death)
}

func testGarbageEnding(t *testing.T, garbage string) {
	j := gameInfoTestJSON() + framesTestJSON() + garbage

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 2, "3rd frame is invalid and should be ignored")
	require.Equal(t, int32(1), frames[0].Turn)
	require.Equal(t, int32(2), frames[1].Turn)
}

func TestReadGameFramesPlusGarbage(t *testing.T) {
	testGarbageEnding(t, "...")
	testGarbageEnding(t, "{")
	testGarbageEnding(t, "{ foo }")
}

func TestReadGameFramesGarbageAfterHeader(t *testing.T) {
	j := gameInfoTestJSON() + "\n\n{\n" + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 2, "garbage should be ignored")
	require.Equal(t, int32(1), frames[0].Turn)
	require.Equal(t, int32(2), frames[1].Turn)
}

func TestReadGameFramesEmpty(t *testing.T) {
	j := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 0)
}

func TestReadGameFramesBlankFile(t *testing.T) {
	openFileReader = fileOpener(map[string]string{
		"myid": "",
	})
	frames, _ := ReadGameFrames("", "myid")

	require.Len(t, frames, 0)
}

func TestReadGameInfoOneLine(t *testing.T) {
	infoJSON := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(infoJSON),
	})
	game, err := ReadGameInfo("", "myid")

	require.NoError(t, err)
	require.Equal(t, "myid", game.ID)
	require.Equal(t, int32(10), game.Width)
	require.Equal(t, int32(12), game.Height)
}

func TestReadGameInfoManyLines(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	game, err := ReadGameInfo("", "myid")

	require.NoError(t, err)
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoManyLinesPlusGarbage(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON() + "asdf"

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	game, err := ReadGameInfo("", "myid")

	require.NoError(t, err, "garbage data after last frame should not break anything")
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoPlusGarbage(t *testing.T) {
	text := gameInfoTestJSON() + "\nfoo\nbar"

	openFileReader = fileOpener(map[string]string{
		"myid": string(text),
	})
	game, err := ReadGameInfo("", "myid")

	require.NoError(t, err, "garbage data after header should not break anything")
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoCorruptJSON(t *testing.T) {
	text := "foo\nbar"

	openFileReader = fileOpener(map[string]string{
		"myid": string(text),
	})
	_, err := ReadGameInfo("", "myid")

	require.NotNil(t, err)
}

func TestReadGameInfoMissingFile(t *testing.T) {
	infoJSON := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(infoJSON),
	})
	_, err := ReadGameInfo("", "wrongid")

	require.NotNil(t, err)
}
