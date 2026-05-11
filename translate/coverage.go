package translate

import (
	"encoding/json"
	"fmt"
	"sort"
)

// CoverageReport describes Java block states that were checked against the
// translator and which states, if any, still fall back to the missing block.
type CoverageReport struct {
	TotalStates int
	Unknown     UnknownReport
}

// UnknownReport tallies Java states that could not be translated.
type UnknownReport struct {
	Counts map[string]int
	Total  int
}

// UnknownStateCount is one row in a sorted unknown-state report.
type UnknownStateCount struct {
	State string
	Count int
}

// TopUnknowns returns the most frequent unknown states in descending order.
func (r CoverageReport) TopUnknowns(limit int) []UnknownStateCount {
	return r.Unknown.Top(limit)
}

// Top returns the most frequent unknown states in descending order.
func (r UnknownReport) Top(limit int) []UnknownStateCount {
	rows := make([]UnknownStateCount, 0, len(r.Counts))
	for state, count := range r.Counts {
		rows = append(rows, UnknownStateCount{State: state, Count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].State < rows[j].State
		}
		return rows[i].Count > rows[j].Count
	})
	if limit >= 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

// JavaStateCoverage enumerates every state in the embedded Java block metadata
// and verifies each state can be translated to a Bedrock palette state.
func JavaStateCoverage() (CoverageReport, error) {
	var javaBlocks []prismaBlock
	if err := json.Unmarshal(javaBlocksJSON, &javaBlocks); err != nil {
		return CoverageReport{}, fmt.Errorf("translate: java_blocks.json: %w", err)
	}

	report := CoverageReport{Unknown: UnknownReport{Counts: map[string]int{}}}
	for _, jb := range javaBlocks {
		propNames, propValues := extractProperties(jb)
		if len(propNames) == 0 {
			report.check(canonicalKey(jb.Name, nil))
			continue
		}
		idx := make([]int, len(propNames))
		for {
			javaProps := make(map[string]string, len(propNames))
			for i, pn := range propNames {
				javaProps[pn] = propValues[i][idx[i]]
			}
			report.check(canonicalKey(jb.Name, javaProps))

			done := true
			for i := len(idx) - 1; i >= 0; i-- {
				idx[i]++
				if idx[i] < len(propValues[i]) {
					done = false
					break
				}
				idx[i] = 0
			}
			if done {
				break
			}
		}
	}
	return report, nil
}

func (r *CoverageReport) check(state string) {
	r.TotalStates++
	res := Lookup(state)
	if res.Recognized {
		return
	}
	key := res.RawKey
	if key == "" {
		key = state
	}
	r.Unknown.Counts[key]++
	r.Unknown.Total++
}
