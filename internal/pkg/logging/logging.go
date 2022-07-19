package logging

import (
	"bufio"
	"fmt"
	. "github.com/hotid/streamsurfer/internal/pkg/structures"
	"os"
	"strconv"
	"time"
)

var logq chan LogMessage

const TimeFormat = "2006-01-02 15:04:05"

type LogMessage struct {
	Severity Severity
	Stream
	Result
}

func LogKeeper(verbose bool, config *Config) {
	var skip error
	var logw *bufio.Writer

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Logger trace:", r)
		}
	}()

	logq = make(chan LogMessage, 1024)
	//logf, skip := os.Create(config.ErrorLog)
	//if skip == nil {
	logw = bufio.NewWriter(os.Stderr)
	//fmt.Printf("Error log: %s\n", config.ErrorLog)
	//} else {
	//	println("Can't create file for error log. Error logging to file skiped.")
	//}

	for {
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(2 * time.Second)
			timeout <- true
		}()

		select {
		case msg := <-logq:
			if skip == nil {
				logw.WriteString(msg.Started.Format(TimeFormat))
				logw.WriteRune(' ')
				switch msg.Severity {
				case WARNING:
					logw.WriteString("warning")
				case ERROR:
					logw.WriteString("error")
				}
				logw.WriteString(": ")
				logw.WriteString(StreamErr2String(msg.Result.ErrType))
				logw.WriteRune(' ')
				logw.WriteString(strconv.Itoa(msg.HTTPCode))
				logw.WriteRune(' ')
				logw.WriteString(strconv.FormatInt(msg.ContentLength, 10))
				logw.WriteRune(' ')
				logw.WriteString(msg.Elapsed.String())
				logw.WriteRune(' ')
				logw.WriteString(msg.Group)
				logw.WriteString(": ")
				logw.WriteString(msg.Name)
				logw.WriteRune('\n')
			}
		case <-timeout:
			if skip == nil {
				_ = logw.Flush()
			}
		}
	}
}

func Log(severity Severity, stream Stream, taskres Result) {
	logq <- LogMessage{Severity: severity, Stream: stream, Result: taskres}
}

// Text representation of stream errors
func StreamErr2String(err ErrType) string {
	switch err {
	case SUCCESS:
		return "success"
	case HLSPARSER:
		return "HLS parser" // debug
	case BADREQUEST:
		return "invalid request" // debug
	case SLOW:
		return "slow response"
	case VERYSLOW:
		return "very slow response"
	case BADSTATUS:
		return "bad status"
	case BADURI:
		return "bad URI"
	case LISTEMPTY: // HLS specific
		return "list empty"
	case BADFORMAT: // HLS specific
		return "bad format"
	case TTLEXPIRED:
		return "TTL expired"
	case RTIMEOUT:
		return "timeout on read"
	case CTIMEOUT:
		return "connection timeout"
	case BADLENGTH:
		return "bad content length value"
	case BODYREAD:
		return "response body error"
	case REFUSED:
		return "connection refused"
	default:
		return "unknown"
	}
}
