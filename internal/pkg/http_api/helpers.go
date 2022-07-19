package http_api

import (
	"encoding/hex"
	"fmt"
	"github.com/hotid/streamsurfer/internal/pkg/structures"
)

func href(url, text string, opts ...string) string {
	switch len(opts) {
	case 1:
		return fmt.Sprintf("<a title=\"%s\" href=\"%s\">%s</a>", opts[0], url, text)
	case 2:
		return fmt.Sprintf("<a title=\"%s\" class=\"%s\" href=\"%s\">%s</a>", opts[0], opts[1], url, text)
	default:
		return fmt.Sprintf("<a href=\"%s\">%s</a>", url, text)
	}
}

func span(text, class string) string {
	return fmt.Sprintf("<span class=\"%s\">%s</span>", class, text)
}

func bytewe(res []byte, err error) []byte {
	return res
}

func KeyFromHex(hexstr string) (structures.Key, error) {
	var key [32]byte
	sum, err := hex.DecodeString(hexstr)
	if err != nil {
		return key, err
	}
	copy(key[:], sum[0:32])
	return key, nil
}
