package torrent

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

type Magnet struct {
	InfoHash    [20]byte
	InfoHashOK  bool
	Trackers    []string
	DisplayName string
}

func ParseMagnet(raw string) (*Magnet, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Scheme) != "magnet" {
		return nil, fmt.Errorf("not a magnet link")
	}

	out := &Magnet{}
	q := u.Query()

	for _, xt := range q["xt"] {
		if ih, ok := parseXt(xt); ok {
			out.InfoHash = ih
			out.InfoHashOK = true
			break
		}
	}
	out.Trackers = append(out.Trackers, q["tr"]...)
	if dn := q.Get("dn"); dn != "" {
		out.DisplayName = dn
	}

	if !out.InfoHashOK {
		return nil, fmt.Errorf("missing or invalid infohash")
	}
	return out, nil
}

func parseXt(xt string) ([20]byte, bool) {
	var zero [20]byte
	xt = strings.ToLower(strings.TrimSpace(xt))
	if !strings.HasPrefix(xt, "urn:btih:") {
		return zero, false
	}
	hash := strings.TrimPrefix(xt, "urn:btih:")
	if len(hash) == 40 {
		b, err := hex.DecodeString(hash)
		if err != nil || len(b) != 20 {
			return zero, false
		}
		var ih [20]byte
		copy(ih[:], b)
		return ih, true
	}
	if len(hash) == 32 {
		b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(hash))
		if err != nil || len(b) != 20 {
			return zero, false
		}
		var ih [20]byte
		copy(ih[:], b)
		return ih, true
	}
	return zero, false
}
