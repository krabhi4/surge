package source

import (
	"net/url"
	"strings"
)

type Kind string

const (
	KindUnknown    Kind = "unknown"
	KindHTTP       Kind = "http"
	KindTorrentURL Kind = "torrent"
	KindMagnet     Kind = "magnet"
)

func Normalize(raw string) string {
	return strings.TrimSpace(raw)
}

func IsHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func IsTorrentURL(raw string) bool {
	if !IsHTTPURL(raw) {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.ToLower(u.Path), ".torrent")
}

func IsMagnet(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if strings.ToLower(u.Scheme) != "magnet" {
		return false
	}
	// Accept any non-empty magnet payload (opaque or query).
	return u.Opaque != "" || u.RawQuery != ""
}

func KindOf(raw string) Kind {
	s := Normalize(raw)
	if s == "" {
		return KindUnknown
	}
	if IsMagnet(s) {
		return KindMagnet
	}
	if IsTorrentURL(s) {
		return KindTorrentURL
	}
	if IsHTTPURL(s) {
		return KindHTTP
	}
	return KindUnknown
}

func IsSupported(raw string) bool {
	return KindOf(raw) != KindUnknown
}

// ParseCommaArg parses a comma-separated input and returns the primary URL and mirrors.
// Mirrors include the primary URL for HTTP/HTTPS inputs (backward compatibility).
func ParseCommaArg(arg string) (string, []string) {
	parts := strings.Split(arg, ",")
	primary := ""
	mirrors := []string{}

	for _, p := range parts {
		clean := strings.TrimSpace(p)
		if clean == "" {
			continue
		}
		if primary == "" {
			if !IsSupported(clean) {
				continue
			}
			primary = clean
			if IsMagnet(primary) {
				mirrors = append(mirrors, primary)
			} else if IsHTTPURL(primary) {
				mirrors = append(mirrors, primary)
			}
			continue
		}
		// Mirrors are HTTP/HTTPS only.
		if IsHTTPURL(clean) {
			mirrors = append(mirrors, clean)
		}
	}

	if primary == "" {
		return "", nil
	}
	return primary, mirrors
}
