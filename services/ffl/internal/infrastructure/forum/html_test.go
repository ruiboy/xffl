package forum

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Real THC post body (br-separated) from the captured 2025 Round 1 page.
const thcBRHTML = `THC - Round 1 <br>
Strap in, it's gonna be a rough ride<br>
<br>
Goals<br>
B King GCS<br>
J Cameron GEE<br>
<br>
Star<br>
N Daicos COL<i data-tag="post_body_end" class="hide"></i>`

// Real Ruiboys post body (table layout), abbreviated, from the same page.
const ruiTableHTML = `My sheepdog is a bit mangy.<br>
<table class="post_content_table"><tbody><tr><td><strong>R1</strong></td>
<td></td>
</tr>
<tr><td><strong>GOALS</strong></td>
<td></td>
</tr>
<tr><td>Jake Waterman - WCE</td>
<td></td>
</tr>
<tr><td><strong>BENCH</strong></td>
<td></td>
</tr>
<tr><td>Levi Ashcroft - Bris</td>
<td>* (INT)</td>
</tr>
<tr><td>Brodie Grundy - Syd</td>
<td>R/H</td>
</tr>
</tbody></table><i data-tag="post_body_end" class="hide"></i>`

func TestPostHTMLToText_BRFormat(t *testing.T) {
	got := PostHTMLToText(thcBRHTML)
	lines := strings.Split(got, "\n")

	// Structural lines survive; the hidden end marker is gone.
	assert.Contains(t, lines, "THC - Round 1")
	assert.Contains(t, lines, "Goals")
	assert.Contains(t, lines, "B King GCS")
	assert.Contains(t, lines, "N Daicos COL")
	assert.NotContains(t, got, "post_body_end")
	assert.NotContains(t, got, "<br>")
}

func TestPostHTMLToText_TableFormat(t *testing.T) {
	got := PostHTMLToText(ruiTableHTML)
	lines := strings.Split(got, "\n")

	assert.Contains(t, lines, "R1")
	assert.Contains(t, lines, "GOALS")
	assert.Contains(t, lines, "Jake Waterman - WCE")
	// A bench row's two cells join with a space, matching the non-table format.
	assert.Contains(t, lines, "Levi Ashcroft - Bris * (INT)")
	assert.Contains(t, lines, "Brodie Grundy - Syd R/H")
	assert.NotContains(t, got, "<td>")
	assert.NotContains(t, got, "<strong>")
}

// The extracted text must feed the parser and yield players.
func TestPostHTMLToText_FeedsParser(t *testing.T) {
	rows, err := NewParser().Parse(t.Context(), "THC", PostHTMLToText(thcBRHTML))
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	assert.Equal(t, "goals", rows[0].Position)
}
