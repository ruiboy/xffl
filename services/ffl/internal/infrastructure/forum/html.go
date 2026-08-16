package forum

import (
	"regexp"
	"strings"
)

var (
	wsRE     = regexp.MustCompile(`\s+`)
	brRE     = regexp.MustCompile(`(?i)<br\s*/?>`)
	trRE     = regexp.MustCompile(`(?i)</tr>`)
	cellRE   = regexp.MustCompile(`(?i)</t[dh]>`)
	tagRE    = regexp.MustCompile(`<[^>]*>`)
	spacesRE = regexp.MustCompile(`[ \t]+`)
)

var entityReplacer = strings.NewReplacer(
	"&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`,
	"&#39;", "'", "&#039;", "'", "&apos;", "'", "&nbsp;", " ",
)

// PostHTMLToText converts a forum post's content HTML (the inner HTML of
// div.content, which uses <br> line breaks and sometimes a <table> layout) into
// plain newline-separated text suitable for Parser.Parse.
//
// Line breaks come from <br> and table row ends (</tr>); table cells (</td>) are
// joined with a space so a "player | code" row reads like the non-table format.
// All other tags are stripped and a few HTML entities decoded.
func PostHTMLToText(html string) string {
	// Whitespace between tags (incl. the source's pretty-print newlines) is
	// insignificant — collapse it so only <br> and </tr> create line breaks.
	s := wsRE.ReplaceAllString(html, " ")
	s = brRE.ReplaceAllString(s, "\n")
	s = trRE.ReplaceAllString(s, "\n")
	s = cellRE.ReplaceAllString(s, " ")
	s = tagRE.ReplaceAllString(s, "")
	s = entityReplacer.Replace(s)

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		out = append(out, strings.TrimSpace(spacesRE.ReplaceAllString(ln, " ")))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
