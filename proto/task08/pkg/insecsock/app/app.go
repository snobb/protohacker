package app

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// HandleLine finds the best item to copy.
func HandleLine(line string) string {
	lines := strings.Split(line, ",")
	maxIdx := 0
	maxVal := -1

	re := regexp.MustCompile(`(^\d+)x`)

	for i, line := range lines {
		mm := re.FindSubmatch([]byte(line))
		if mm == nil || len(mm) < 2 {
			log.Printf("no numbers found")
			continue
		}

		val, err := strconv.Atoi(string(mm[1]))
		if err != nil {
			log.Printf("cannot parse number")
			continue
		}

		if val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}

	return lines[maxIdx]
}
