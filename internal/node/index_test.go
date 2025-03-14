package node

import (
	"testing"
	"time"
)

func TestSortIndex(t *testing.T) {
	latestDate := time.Date(2024, time.April, 11, 0, 0, 0, 0, time.UTC)

	idx := []IndexEntry{
		{
			Date: time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			Date: time.Date(2022, time.April, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			Date: latestDate,
		},
		{
			Date: time.Date(2020, time.April, 11, 0, 0, 0, 0, time.UTC),
		},
	}

	sortIndex(&idx)

	if !idx[0].Date.Equal(latestDate) {
		t.Errorf("first index isn't the latest (expected %s, got %s)", idx[0].Date.String(), latestDate.String())
	}

	for i := range len(idx) - 1 {
		a := idx[i]
		b := idx[i+1]

		if a.Date.Before(b.Date) {
			t.Error("index not sorted")
		}
	}
}
