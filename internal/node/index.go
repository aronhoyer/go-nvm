package node

import (
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

type IndexEntry struct {
	Version, NPM, LTS string
	Date              time.Time
}

func GetRemoteIndex() ([]IndexEntry, error) {
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

	sortIndex(&idx)

	return idx, nil
}

func sortIndex(idx *[]IndexEntry) {
	slices.SortFunc(*idx, func(a IndexEntry, b IndexEntry) int {
		if a.Date.After(b.Date) {
			return -1
		}

		if a.Date.Before(b.Date) {
			return 1
		}

		return 0
	})
}

func parseIndexLine(line string) (IndexEntry, error) {
	// version	date	files	npm	v8	uv	zlib	openssl	modules	lts	security
	parts := strings.Fields(line)

	version, npm, lts := parts[0], parts[3], parts[9]

	date, err := time.Parse("2006-01-02", parts[1])
	if err != nil {
		return IndexEntry{}, err
	}

	if npm == "-" {
		npm = ""
	}

	if lts == "-" {
		lts = ""
	}

	return IndexEntry{version, npm, lts, date}, nil
}
