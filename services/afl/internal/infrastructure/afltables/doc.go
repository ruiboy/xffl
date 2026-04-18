// Package afltables fetches AFL player match statistics from afltables.com
// and writes them into the afl domain via the StatsProvider port.
//
// Cache policy: data is fetched at most once per week. The cache is
// invalidated on Monday to pick up the previous weekend's results.
package afltables
