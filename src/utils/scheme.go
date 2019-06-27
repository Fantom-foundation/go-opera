package utils

import (
	"bufio"
	"strings"
)

// TextColumns join side-by-side.
func TextColumns(texts ...string) string {
	var (
		columns = make([][]string, len(texts))
		widthes = make([]int, len(texts))
	)

	for i, text := range texts {
		scanner := bufio.NewScanner(strings.NewReader(text))
		for scanner.Scan() {
			line := scanner.Text()
			columns[i] = append(columns[i], line)
			if widthes[i] < len([]rune(line)) {
				widthes[i] = len([]rune(line))
			}
		}
	}

	var (
		res strings.Builder
		j   int
	)
	for {
		eof := true
		for i := range columns {
			var s string
			if len(columns[i]) > j {
				s = columns[i][j]
				s = s + strings.Repeat(" ", widthes[i]-len([]rune(s)))
				eof = false
			} else {
				s = strings.Repeat(" ", widthes[i])
			}
			res.WriteString(s)
			res.WriteString("\t")
		}
		res.WriteString("\n")
		j++
		if eof {
			break
		}
	}

	return res.String()
}
