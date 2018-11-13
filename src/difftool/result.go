package difftool

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andrecronje/lachesis/src/node"
)

// Diff contains and prints differences details
type Diff struct {
	Err error `json:"-"`

	node            [2]*node.Node `json:"-"`
	IDs             [2]int64
	BlocksGap       int64 `json:",omitempty"`
	FirstBlockIndex int64 `json:",omitempty"`
	RoundGap        int64 `json:",omitempty"`
	FirstRoundIndex int64 `json:",omitempty"`

	Descr string `json:"-"`
}

// Result is a set of differences
type Result []*Diff

/*
 * Diff's methods
 */

func (d *Diff) IsEmpty() bool {
	has := d.FirstBlockIndex > 0 || d.FirstRoundIndex > 0
	return !has
}

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

func (d *Diff) AddDescr(s string) {
	d.Descr = d.Descr + s + "\n"
}

/*
 * Result's methods
 */

func (r Result) IsEmpty() bool {
	for _, diff := range r {
		if !diff.IsEmpty() {
			return false
		}
	}
	return true
}

func (r Result) ToString() string {
	var output []string
	for _, diff := range r {
		if !diff.IsEmpty() {
			output = append(output, diff.ToString())
			if diff.Descr != "" {
				output = append(output, "\t"+diff.Descr)
			}
		}
	}
	return strings.Join(output, "\n")
}
