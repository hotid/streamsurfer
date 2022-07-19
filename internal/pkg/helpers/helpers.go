package helpers

import (
	"fmt"
	. "github.com/hotid/streamsurfer/internal/pkg/structures"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func FullPath(name string) string {
	return os.ExpandEnv(strings.Replace(name, "~", os.Getenv("HOME"), 1))
}

func NewTimeoutClient(args ...interface{}) *http.Client {
	// Default configuration
	config := &HTTPConfig{
		ConnectTimeout:   1 * time.Second,
		ReadWriteTimeout: 1 * time.Second,
	}

	// merge the default with user input if there is one
	if len(args) == 1 {
		timeout := args[0].(time.Duration)
		config.ConnectTimeout = timeout
		config.ReadWriteTimeout = timeout
	}

	if len(args) == 2 {
		config.ConnectTimeout = args[0].(time.Duration)
		config.ReadWriteTimeout = args[1].(time.Duration)
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}

func TimeoutDialer(config *HTTPConfig) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		tcp_conn := conn.(*net.TCPConn)
		tcp_conn.SetLinger(0) // because we have mass connects/disconnects
		return conn, nil
	}
}

// Returns proper user agent string for the HTTP-headers.
func UserAgent(cfg *Config) string {
	if len(cfg.UserAgents) > 0 {
		return cfg.UserAgents[rand.Intn(len(cfg.UserAgents)-1)]
	} else {
		return fmt.Sprintf("%s/%s", "hls-monitor", "0.0.1a")
	}
}

func String2StreamErr(s string) ErrType {
	switch strings.ToLower(s) {
	case "success":
		return SUCCESS
	case "debug":
		return DEBUG_LEVEL
	case "hlsparser":
		return HLSPARSER
	case "badrequest":
		return BADREQUEST
	case "warning":
		return WARNING_LEVEL
	case "slow":
		return SLOW
	case "veryslow":
		return VERYSLOW
	case "badstatus":
		return BADSTATUS
	case "baduri":
		return BADURI
	case "listempty": // HLS specific
		return LISTEMPTY
	case "badformat": // HLS specific
		return BADFORMAT
	case "ttlexpired":
		return TTLEXPIRED
	case "rtimeout":
		return RTIMEOUT
	case "error":
		return ERROR_LEVEL
	case "ctimeout":
		return CTIMEOUT
	case "badlength":
		return BADLENGTH
	case "bodyread":
		return BODYREAD
	case "critical":
		return CRITICAL_LEVEL
	case "refused":
		return REFUSED
	default:
		return UNKERR
	}
}

func StreamType2String(t StreamType) string {
	switch t {
	case SAMPLE:
		return "sample"
	case HLS:
		return "hls"
	case HDS:
		return "hds"
	case WV:
		return "wv"
	case HTTP:
		return "http"
	default:
		return "unknown"
	}
}
