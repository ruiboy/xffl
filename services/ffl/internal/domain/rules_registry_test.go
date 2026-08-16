package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AllRules lists every era, ascending by id, so the builder can offer them.
func TestAllRules(t *testing.T) {
	all := AllRules()
	require.Len(t, all, 5)

	ids := make([]string, len(all))
	for i, r := range all {
		ids[i] = r.ID
	}
	assert.Equal(t, []string{"1998", "1999", "2000", "2001", "2011"}, ids, "sorted ascending by id")
}

// Describe summarises what distinguishes an era, for the rules dropdown.
func TestRulesDescribe(t *testing.T) {
	d2011 := rules2011.Describe()
	assert.Contains(t, d2011, "goals 5")
	assert.Contains(t, d2011, "tackles 4")
	assert.Contains(t, d2011, "bench 4")
	assert.Contains(t, strings.ToLower(d2011), "interchange")

	d2000 := rules2000.Describe()
	assert.Contains(t, d2000, "tackles 4")
	assert.Contains(t, strings.ToLower(d2000), "no bench")
	assert.NotContains(t, strings.ToLower(d2000), "hitout")

	// Pre-1999 the star also scores hitouts — the one composition quirk worth surfacing.
	d1998 := rules1998.Describe()
	assert.Contains(t, d1998, "tackles 3")
	assert.Contains(t, strings.ToLower(d1998), "hitout")
}
