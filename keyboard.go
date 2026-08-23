package telejoon

import "slices"

// isEmptyString reports break markers in reply keyboard button lists.
func isEmptyString(s string) bool { return s == "" }

// isNilRow reports break markers in inline keyboard row lists.
func isNilRow(m map[string]string) bool { return m == nil }

// chunkIntoRows splits a flat item list into keyboard rows.
//
// Break markers ("" for reply buttons, nil for inline rows, emitted for the
// breakBefore/breakAfter button options) close the current row without being
// rendered themselves — passing them through to the underlying library would
// render empty KeyboardButtons / empty rows, which Telegram rejects.
//
// Row sizes come from formation; once the formation is exhausted its last
// entry is repeated (instead of falling back to one item per row). Without a
// formation, maxPerRow applies; a non-positive limit means one item per row.
func chunkIntoRows[T any](
	items []T,
	isBreakMarker func(T) bool,
	maxPerRow int,
	formation []int,
	reverse bool,
) [][]T {

	var rows [][]T
	var row []T

	flush := func() {
		if len(row) == 0 {
			return
		}

		if reverse {
			slices.Reverse(row)
		}

		rows = append(rows, row)
		row = nil
	}

	for _, item := range items {
		if isBreakMarker(item) {
			flush()
			continue
		}

		row = append(row, item)

		if len(row) >= rowSize(len(rows), maxPerRow, formation) {
			flush()
		}
	}

	flush()

	return rows
}

func rowSize(rowIndex, maxPerRow int, formation []int) int {
	if rowIndex < len(formation) {
		if formation[rowIndex] > 0 {
			return formation[rowIndex]
		}
	} else if len(formation) > 0 {
		if last := formation[len(formation)-1]; last > 0 {
			return last
		}
	}

	if maxPerRow > 0 {
		return maxPerRow
	}

	return 1
}
