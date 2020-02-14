package rules

import (
	"bytes"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

var snakeURL = "http://good-server"

const validationCount = 4

func TestValidatePing404(t *testing.T) {
	createClient = validateMockClient(t, snakeURL+"/ping", "{}", 404, 200)
	response := ValidatePing("1234", snakeURL, 200)

	errorZero := response.Errors[0]
	require.Equal(t, errorZero, "incorrect http response code, got 404, expected 200")
}

func TestValidateTrailingSlash(t *testing.T) {
	createClient = validateMockClient(t, snakeURL+"/ping", "{}", 200, 200)
	response := ValidatePing("1234", snakeURL+"/", 200)
	require.Equal(t, int32(validationCount), response.GetScore().GetChecksPassed())
}

func TestValidatePing500(t *testing.T) {
	createClient = validateMockClient(t, snakeURL+"/ping", "{}", 500, 200)
	response := ValidatePing("1234", snakeURL, 200)

	errorZero := response.Errors[0]
	require.Equal(t, errorZero, "incorrect http response code, got 500, expected 200")
}

func TestValidateEnd(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		StatusCode: 200,
		Raw:        "",
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateEnd, snakeURL+"/end", expected)
}
func TestValidateMove(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "{  }",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateMove, snakeURL+"/move", expected)
}
func TestValidatePing(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "{  }",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidatePing, snakeURL+"/ping", expected)
}

func TestValidateStart(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "{ \"color\": \"blue\" }",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateStart, snakeURL+"/start", expected)
}

func TestValidateStartBadJson(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Bad response format",
		Raw:        "{ color\": \"blue\" }",
		StatusCode: 200,
		Errors:     []string{"invalid character 'c' looking for beginning of object key string"},
		Score: &pb.Score{
			ChecksPassed: validationCount - 1,
			ChecksFailed: 1,
		},
	}
	validateWithJSON(t, ValidateStart, snakeURL+"/start", expected)
}

func TestValidateEndBadJsonIsOK(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "WE DON'T CARE ABOUT THE RESPONSE FORMAT",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateEnd, snakeURL+"/end", expected)
}
func TestValidatePingBadJsonIsOK(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "WE DON'T CARE ABOUT THE RESPONSE FORMAT",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: validationCount,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidatePing, snakeURL+"/ping", expected)
}
func TestValidateStartBadUrl(t *testing.T) {
	response := ValidateStart("1234", "start", 100)
	require.True(t, strings.Contains(response.Message, "Snake URL not valid"), response.Message)
	require.Equal(t, []string{"invalid url 'start'"}, response.Errors)
}
func TestValidateSlowUrl(t *testing.T) {
	slowSnake := int32(1)
	createClient = validateMockClient(t, snakeURL+"/move", "{}", 200, time.Duration(slowSnake+1)*time.Millisecond)
	response := ValidateMove("1234", snakeURL, slowSnake)
	errorZero := response.Errors[0]
	var digitsRegexp = regexp.MustCompile(`snake took (\d+) ms`)
	digitString := digitsRegexp.FindStringSubmatch(errorZero)
	digits, _ := strconv.Atoi(digitString[1])
	assert.True(t, int32(digits) > slowSnake, "Should have found larger amount than the slow snake limit")
}

func validateWithJSON(t *testing.T, status func(id string, url string, slowSnakeMS int32) *pb.SnakeResponseStatus, url string, expected *pb.SnakeResponseStatus) {
	createClient = validateMockClient(t, url, expected.Raw, 200, 0)
	response := status("1234", snakeURL, 100)
	require.True(t, strings.Contains(response.Message, expected.Message), "got: "+response.Message+", expected: "+expected.Message)
	require.Equal(t, expected.Errors, response.Errors)
	require.Equal(t, expected.Raw, response.Raw)
	require.Equal(t, expected.Score.ChecksPassed, response.Score.ChecksPassed, "Passed count mismatch")
	require.Equal(t, expected.Score.ChecksFailed, response.Score.ChecksFailed, "Failed count mismatch")
}

func validateMockClient(t *testing.T, url, bodyJSON string, statusCode int, sleep time.Duration) func(time.Duration) httpClient {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString(bodyJSON)

	return func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(reqURL string) *http.Response {
				time.Sleep(sleep)
				if reqURL != url {
					require.Fail(t, "invalid url")
				}
				return &http.Response{
					Body:       body,
					StatusCode: statusCode,
				}
			},
		}
	}
}
