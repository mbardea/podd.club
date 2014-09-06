package util

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/mbardea/podd.club/logger"
)

func ParseRangeHeader(headers *http.Header, origStartPos int64, origEndPos int64) (int64, int64) {
	startPos := origStartPos
	endPos := origEndPos
	// Check HTTP header in case it was a range request
	rangeHeader := headers.Get("Range")
	if rangeHeader != "" {
		// Example: "Range: [bytes=0-]"
		reParser := regexp.MustCompile(".*\\[bytes=([0-9]+)-([0-9]*)\\].*")
		match := reParser.FindStringSubmatch(rangeHeader)
		if len(match) > 0 {
			strStartPos := match[1]
			strEndPos := match[2]
			startPos, _ = strconv.ParseInt(strStartPos, 10, 64)
			if len(strEndPos) > 0 {
				endPos, _ = strconv.ParseInt(strEndPos, 10, 64)
			}
			logger.Infof("Parsed range start: %d", startPos)
		}
	}
	if endPos < startPos {
		endPos = startPos
	}

	return startPos, endPos
}
