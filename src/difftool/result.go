package difftool

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/node"
)

// Diff contains and prints differences details
type Diff struct {
	Err error `json:"-"`

	node            [2]*node.Node `json:"-"`
	IDs             [2]uint64
	BlocksGap       int64 `json:",omitempty"`
	FirstBlockIndex int64 `json:",omitempty"`
	RoundGap        int64 `json:",omitempty"`
	FirstRoundIndex int64 `json:",omitempty"`

	Description string `json:"-"`
}

// Result is a set of differences
type Result []*Diff

/*
 * Diff's methods
 */

// IsEmpty is the diff empty
func (d *Diff) IsEmpty() bool {
	has := d.FirstBlockIndex > 0 || d.FirstRoundIndex > 0
	return !has
}

// ToString converts to a string
func (d *Diff) ToString() string {
	if d.Err != nil {
		return fmt.Sprintf("ERR: %s", d.Err.Error())
	}
	if d.IsEmpty() {
		return ""
	}

	raw, err := json.Marshal(d)
	if err != nil {
		return fmt.Sprintf("JSON: %s", err.Error())
	}
	return string(raw)
}

// AddDescription appends description to the diff
func (d *Diff) AddDescription(s string) {
	d.Description = d.Description + s + "\n"
}

/*
 * Result's methods
 */

// IsEmpty is the result empty
func (r Result) IsEmpty() bool {
	for _, diff := range r {
		if !diff.IsEmpty() {
			return false
		}
	}
	return true
}

// ToString result to string
func (r Result) ToString() string {
	var output []string
	for _, diff := range r {
		if !diff.IsEmpty() {
			output = append(output, diff.ToString())
			if diff.Description != "" {
				output = append(output, "\t"+diff.Description)
			}
		}
	}
	return strings.Join(output, "\n")
}
