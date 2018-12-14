package cron

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry_ByTimeAsc_Sort(t *testing.T) {
	expressions := []string{
		"@every 1h30m",
		"@every 2h30m",
		"@every 3h30m",
		"@every 30m",
		"@every 1h",
	}

	var entries []*Entry
	for i, exp := range expressions {
		entries = append(entries, &Entry{Expression: exp, ID: strconv.FormatInt(int64(i), 10)})
	}

	sort.Sort(ByTimeAsc(entries))

	wantedOrder := []string{
		"@every 30m",
		"@every 1h",
		"@every 1h30m",
		"@every 2h30m",
		"@every 3h30m",
	}

	for i, e := range entries {
		assert.Equal(t, wantedOrder[i], e.Expression)
	}
}
