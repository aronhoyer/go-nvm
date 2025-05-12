package node

import (
	"io"
	"net/http"
	"os"
	"strings"
)

type IndexEntry struct {
	Version, LTS string
}

func GetRemoteIndex() ([]IndexEntry, error) {
	// Although Node ships a JSON distro index, we prefer TSV for a few reasons. Chief among them being that parsing
	// TSV is about **10 times faster**. Literally.
	//
	// In controlled benchmarks, the below TSV parser cleared 1000 iterations in ~200ms, whereas the equivalent JSON
	// parsing (using encoding/json) dragged on for ~2 entire seconds (same number of iterations, same data).
	//
	// So yeah... just don't parse JSON. It's slow, it's bloated, and it'll waste your CPU cycles validating curly
	// braces like it's the highlight of your runtime.
	//
	// Additionally, and get this, the `lts` field in Node's JSON distribution index can be **either** boolean or
	// string. Slowing down JSON parsing EVEN FURTHER.
	//
	// Long live raw data.

	res, err := http.Get("https://nodejs.org/dist/index.tab")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	idxLines := strings.Split(strings.TrimSpace(string(b)), "\n")
	var idx []IndexEntry

	for i := 1; i < len(idxLines); i++ {
		entry, err := parseIndexLine(idxLines[i])
		if err != nil {
			return nil, err
		}

		idx = append(idx, entry)
	}

	return idx, nil
}

func GetLocalIndex(idxPath string) ([]IndexEntry, error) {
	entries, err := os.ReadDir(idxPath)
	if err != nil {
		return nil, err
	}

	var idxEntries []IndexEntry

	for _, entry := range entries {
		idxEntries = append(idxEntries, IndexEntry{Version: entry.Name()})
	}

	return idxEntries, nil
}

func parseIndexLine(line string) (IndexEntry, error) {
	// version	date	files	npm	v8	uv	zlib	openssl	modules	lts	security
	parts := strings.Fields(line)

	version, lts := parts[0], parts[9]

	if lts == "-" {
		lts = ""
	}

	return IndexEntry{version, lts}, nil
}
