package torrent

import (
	"crypto/sha1"
	"fmt"

	"github.com/surge-downloader/surge/internal/torrent/bencode"
)

func ParseTorrent(data []byte) (*TorrentMeta, error) {
	rootAny, infoBytes, err := bencode.DecodeWithSpanKey(data, "info")
	if err != nil {
		return nil, err
	}
	root, ok := rootAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid torrent root")
	}
	infoAny, ok := root["info"]
	if !ok {
		return nil, fmt.Errorf("missing info dict")
	}
	infoMap, ok := infoAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid info dict")
	}

	info, err := parseInfo(infoMap)
	if err != nil {
		return nil, err
	}

	if len(infoBytes) == 0 {
		encoded, encErr := bencode.Encode(infoMap)
		if encErr != nil {
			return nil, encErr
		}
		infoBytes = encoded
	}
	hash := sha1.Sum(infoBytes)

	meta := &TorrentMeta{
		Info:      info,
		InfoHash:  hash,
		InfoBytes: infoBytes,
	}

	if v, ok := root["announce"]; ok {
		if s, ok := v.([]byte); ok {
			meta.Announce = string(s)
		}
	}
	if v, ok := root["announce-list"]; ok {
		meta.AnnounceList = parseAnnounceList(v)
	}

	return meta, nil
}

func parseInfo(m map[string]any) (Info, error) {
	var info Info

	if v, ok := m["name"]; ok {
		if s, ok := v.([]byte); ok {
			info.Name = string(s)
		}
	}
	if v, ok := m["piece length"]; ok {
		if n, ok := v.(int64); ok {
			info.PieceLength = n
		}
	}
	if v, ok := m["pieces"]; ok {
		if b, ok := v.([]byte); ok {
			info.Pieces = b
		}
	}
	if v, ok := m["length"]; ok {
		if n, ok := v.(int64); ok {
			info.Length = n
		}
	}
	if v, ok := m["files"]; ok {
		files, err := parseFiles(v)
		if err != nil {
			return info, err
		}
		info.Files = files
	}

	if info.PieceLength == 0 || len(info.Pieces) == 0 || info.Name == "" {
		return info, fmt.Errorf("invalid info dict")
	}
	if info.Length == 0 && len(info.Files) == 0 {
		return info, fmt.Errorf("missing length/files")
	}
	return info, nil
}

func parseFiles(v any) ([]FileEntry, error) {
	list, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid files list")
	}
	files := make([]FileEntry, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid file entry")
		}
		var fe FileEntry
		if n, ok := m["length"].(int64); ok {
			fe.Length = n
		}
		if p, ok := m["path"]; ok {
			pathList, ok := p.([]any)
			if !ok {
				return nil, fmt.Errorf("invalid file path")
			}
			for _, part := range pathList {
				b, ok := part.([]byte)
				if !ok {
					return nil, fmt.Errorf("invalid path element")
				}
				fe.Path = append(fe.Path, string(b))
			}
		}
		if fe.Length <= 0 || len(fe.Path) == 0 {
			return nil, fmt.Errorf("invalid file entry data")
		}
		files = append(files, fe)
	}
	return files, nil
}

func parseAnnounceList(v any) [][]string {
	var out [][]string
	list, ok := v.([]any)
	if !ok {
		return out
	}
	for _, tierAny := range list {
		tierList, ok := tierAny.([]any)
		if !ok {
			continue
		}
		var tier []string
		for _, t := range tierList {
			if b, ok := t.([]byte); ok {
				tier = append(tier, string(b))
			}
		}
		if len(tier) > 0 {
			out = append(out, tier)
		}
	}
	return out
}
