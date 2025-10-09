// Copied from https://github.com/sergi/go-diff
// Customized by Takuto Nagami (logica)
package main

import (
	"strings"
	"time"
)

// Operation defines the operation of a diff item.
type Operation int8

const (
	// DiffDelete item represents a delete diff.
	DiffDelete Operation = -1
	// DiffInsert item represents an insert diff.
	DiffInsert Operation = 1
	// DiffEqual item represents an equal diff.
	DiffEqual Operation = 0
)

// Diff represents one diff operation
type Diff struct {
	Type Operation
	Text string
}

// splice removes amount elements from slice at index index, replacing them with elements.
func splice(slice []Diff, index int, amount int, elements ...Diff) []Diff {
	if len(elements) == amount {
		// Easy case: overwrite the relevant items.
		copy(slice[index:], elements)

		return slice
	}

	if len(elements) < amount {
		// Fewer new items than old.
		// Copy in the new items.
		copy(slice[index:], elements)
		// Shift the remaining items left.
		copy(slice[index+len(elements):], slice[index+amount:])
		// Calculate the new end of the slice.
		end := len(slice) - amount + len(elements)
		// Zero stranded elements at end so that they can be garbage collected.
		tail := slice[end:]
		for i := range tail {
			tail[i] = Diff{}
		}

		return slice[:end]
	}
	// More new items than old.
	// Make room in slice for new elements.
	// There's probably an even more efficient way to do this,
	// but this is simple and clear.
	need := len(slice) - amount + len(elements)
	for len(slice) < need {
		slice = append(slice, Diff{})
	}
	// Shift slice elements right to make room for new elements.
	copy(slice[index+len(elements):], slice[index+amount:])
	// Copy in new elements.
	copy(slice[index:], elements)

	return slice
}

// DiffLinesToRunes splits two texts into a list of runes.
func DiffLinesToRunes(text1, text2 string) ([]rune, []rune, []string) {
	chars1, chars2, lineArray := diffLinesToRunes(text1, text2)

	return chars1, chars2, lineArray
}

// diffLinesToRunes splits two texts into a list of strings. Each string represents one line.
func diffLinesToRunes(text1, text2 string) ([]rune, []rune, []string) {
	// '\x00' is a valid character, but various debuggers don't like it. So we'll insert a junk entry to avoid generating a null character.
	lineArray := []string{""} // e.g. lineArray[4] == 'Hello\n'

	lineHash := make(map[string]int)
	// Each string has the index of lineArray which it points to
	strIndexArray1 := diffLinesToRunesMunge(text1, &lineArray, lineHash)
	strIndexArray2 := diffLinesToRunesMunge(text2, &lineArray, lineHash)

	return strIndexArray1, strIndexArray2, lineArray
}

// diffLinesToRunesMunge splits a text into an array of strings, and reduces the texts to a []string.
func diffLinesToRunesMunge(text string, lineArray *[]string, lineHash map[string]int) []rune {
	// Walk the text, pulling out a substring for each line. text.split('\n') would temporarily double our memory footprint.
	// Modifying text would create many large strings to garbage collect.
	lineStart := 0
	lineEnd := -1
	str := []rune{}

	for lineEnd < len(text)-1 {
		lineEnd = indexOf(text, "\n", lineStart)

		if lineEnd == -1 {
			lineEnd = len(text) - 1
		}

		line := text[lineStart : lineEnd+1]
		lineStart = lineEnd + 1
		lineValue, ok := lineHash[line]

		if ok {
			str = append(str, rune(lineValue))
		} else {
			*lineArray = append(*lineArray, line)
			lineHash[line] = len(*lineArray) - 1
			str = append(str, rune(len(*lineArray)-1))
		}
	}

	return str
}

// DiffMainRunes finds the differences between two rune sequences.
// If an invalid UTF-8 sequence is encountered, it will be replaced by the Unicode replacement character.
func DiffMainRunes(text1, text2 []rune) []Diff {
	deadline := time.Now().Add(time.Second)

	return diffMainRunes(text1, text2, deadline)
}

func diffMainRunes(text1, text2 []rune, deadline time.Time) []Diff {
	if runesEqual(text1, text2) {
		var diffs []Diff
		if len(text1) > 0 {
			diffs = append(diffs, Diff{DiffEqual, string(text1)})
		}

		return diffs
	}
	// Trim off common prefix (speedup).
	commonLength := commonPrefixLength(text1, text2)
	commonPrefix := text1[:commonLength]
	text1 = text1[commonLength:]
	text2 = text2[commonLength:]

	// Trim off common suffix (speedup).
	commonLength = commonSuffixLength(text1, text2)
	commonSuffix := text1[len(text1)-commonLength:]
	text1 = text1[:len(text1)-commonLength]
	text2 = text2[:len(text2)-commonLength]

	// Compute the diff on the middle block.
	diffs := diffCompute(text1, text2, deadline)

	// Restore the prefix and suffix.
	if len(commonPrefix) != 0 {
		diffs = append([]Diff{{DiffEqual, string(commonPrefix)}}, diffs...)
	}

	if len(commonSuffix) != 0 {
		diffs = append(diffs, Diff{DiffEqual, string(commonSuffix)})
	}

	return DiffCleanupMerge(diffs)
}

// commonPrefixLength returns the length of the common prefix of two rune slices.
func commonPrefixLength(text1, text2 []rune) int {
	// Linear search. See comment in commonSuffixLength.
	length := 0
	for ; length < len(text1) && length < len(text2); length++ {
		if text1[length] != text2[length] {
			return length
		}
	}

	return length
}

// commonSuffixLength returns the length of the common suffix of two rune slices.
func commonSuffixLength(text1, text2 []rune) int {
	// Use linear search rather than the binary search discussed at https://neil.fraser.name/news/2007/10/09/.
	// See discussion at https://github.com/sergi/go-diff/issues/54.
	index1 := len(text1)
	index2 := len(text2)

	for n := 0; ; n++ {
		index1--
		index2--

		if index1 < 0 || index2 < 0 || text1[index1] != text2[index2] {
			return n
		}
	}
}

// diffCompute finds the differences between two rune slices.  Assumes that the texts do not have any common prefix or suffix.
func diffCompute(text1, text2 []rune, deadline time.Time) []Diff {
	diffs := []Diff{}
	if len(text1) == 0 {
		// Just add some text (speedup).
		return append(diffs, Diff{DiffInsert, string(text2)})
	} else if len(text2) == 0 {
		// Just delete some text (speedup).
		return append(diffs, Diff{DiffDelete, string(text1)})
	}

	var longText, shortText []rune
	if len(text1) > len(text2) {
		longText = text1
		shortText = text2
	} else {
		longText = text2
		shortText = text1
	}

	if i := runesIndex(longText, shortText); i != -1 {
		operation := DiffInsert
		// Swap insertions for deletions if diff is reversed.
		if len(text1) > len(text2) {
			operation = DiffDelete
		}
		// Shorter text is inside the longer text (speedup).
		return []Diff{
			{operation, string(longText[:i])},
			{DiffEqual, string(shortText)},
			{operation, string(longText[i+len(shortText):])},
		}
	} else if len(shortText) == 1 {
		// Single character string.
		// After the previous speedup, the character can't be an equality.
		return []Diff{
			{DiffDelete, string(text1)},
			{DiffInsert, string(text2)},
		}
	} else if halfMatch := diffHalfMatch(text1, text2); halfMatch != nil { // Check to see if the problem can be split in two.
		// A half-match was found, sort out the return data.
		text1A := halfMatch[0]
		text1B := halfMatch[1]
		text2A := halfMatch[2]
		text2B := halfMatch[3]
		midCommon := halfMatch[4]
		// Send both pairs off for separate processing.
		diffsA := diffMainRunes(text1A, text2A, deadline)
		diffsB := diffMainRunes(text1B, text2B, deadline)
		// Merge the results.
		diffs := diffsA
		diffs = append(diffs, Diff{DiffEqual, string(midCommon)})
		diffs = append(diffs, diffsB...)

		return diffs
	}

	return diffBisect(text1, text2, deadline)
}

// DiffBisect finds the 'middle snake' of a diff, split the problem in two and return the recursively constructed diff.
// If an invalid UTF-8 sequence is encountered, it will be replaced by the Unicode replacement character.
// See Myers 1986 paper: An O(ND) Difference Algorithm and Its Variations.
func DiffBisect(text1, text2 string, deadline time.Time) []Diff {
	// Unused in this code, but retained for interface compatibility.
	return diffBisect([]rune(text1), []rune(text2), deadline)
}

// diffBisect finds the 'middle snake' of a diff, splits the problem in two and returns the recursively constructed diff.
// See Myers's 1986 paper: An O(ND) Difference Algorithm and Its Variations.
// nolint: gocognit, gocyclo, cyclop, funlen
func diffBisect(runes1, runes2 []rune, deadline time.Time) []Diff {
	// Cache the text lengths to prevent multiple calls.
	runes1Len, runes2Len := len(runes1), len(runes2)

	maxD := (runes1Len + runes2Len + 1) / 2 // nolint: mnd
	vOffset := maxD
	vLength := 2 * maxD // nolint: mnd

	v1 := make([]int, vLength) // nolint: varnamelen
	v2 := make([]int, vLength) // nolint: varnamelen

	for i := range v1 {
		v1[i] = -1
		v2[i] = -1
	}

	v1[vOffset+1] = 0
	v2[vOffset+1] = 0

	delta := runes1Len - runes2Len
	// If the total number of characters is odd, then the front path will collide with the reverse path.
	front := (delta%2 != 0)
	// Offsets for start and end of k loop. Prevents mapping of space beyond the grid.
	k1start := 0
	k1end := 0
	k2start := 0
	k2end := 0

	for d := range maxD { // nolint: varnamelen
		// Bail out if deadline is reached.
		if !deadline.IsZero() && d%16 == 0 && time.Now().After(deadline) {
			break
		}

		// Walk the front path one step.
		for k1 := -d + k1start; k1 <= d-k1end; k1 += 2 { // nolint: varnamelen
			k1Offset := vOffset + k1

			var x1 int // nolint: varnamelen
			if k1 == -d || (k1 != d && v1[k1Offset-1] < v1[k1Offset+1]) {
				x1 = v1[k1Offset+1]
			} else {
				x1 = v1[k1Offset-1] + 1
			}

			y1 := x1 - k1 // nolint: varnamelen
			for x1 < runes1Len && y1 < runes2Len {
				if runes1[x1] != runes2[y1] {
					break
				}

				x1++
				y1++
			}

			v1[k1Offset] = x1
			switch {
			case x1 > runes1Len:
				// Ran off the right of the graph.
				k1end += 2

			case y1 > runes2Len:
				// Ran off the bottom of the graph.
				k1start += 2

			case front:
				k2Offset := vOffset + delta - k1
				if k2Offset >= 0 && k2Offset < vLength && v2[k2Offset] != -1 {
					// Mirror x2 onto top-left coordinate system.
					x2 := runes1Len - v2[k2Offset]
					if x1 >= x2 {
						// Overlap detected.
						return diffBisectSplit(runes1, runes2, x1, y1, deadline)
					}
				}
			}
		}
		// Walk the reverse path one step.
		for k2 := -d + k2start; k2 <= d-k2end; k2 += 2 { // nolint: varnamelen
			k2Offset := vOffset + k2

			var x2 int // nolint: varnamelen
			if k2 == -d || (k2 != d && v2[k2Offset-1] < v2[k2Offset+1]) {
				x2 = v2[k2Offset+1]
			} else {
				x2 = v2[k2Offset-1] + 1
			}

			y2 := x2 - k2 // nolint: varnamelen
			for x2 < runes1Len && y2 < runes2Len {
				if runes1[runes1Len-x2-1] != runes2[runes2Len-y2-1] {
					break
				}

				x2++
				y2++
			}

			v2[k2Offset] = x2
			switch {
			case x2 > runes1Len:
				// Ran off the left of the graph.
				k2end += 2

			case y2 > runes2Len:
				// Ran off the top of the graph.
				k2start += 2

			case !front:
				k1Offset := vOffset + delta - k2
				if k1Offset >= 0 && k1Offset < vLength && v1[k1Offset] != -1 {
					x1 := v1[k1Offset] // nolint: varnamelen
					y1 := vOffset + x1 - k1Offset
					// Mirror x2 onto top-left coordinate system.
					x2 = runes1Len - x2
					if x1 >= x2 {
						// Overlap detected.
						return diffBisectSplit(runes1, runes2, x1, y1, deadline)
					}
				}
			}
		}
	}
	// Diff took too long and hit the deadline or number of diffs equals number of characters, no commonality at all.
	return []Diff{
		{DiffDelete, string(runes1)},
		{DiffInsert, string(runes2)},
	}
}

func diffBisectSplit(runes1, runes2 []rune, x, y int, deadline time.Time) []Diff {
	runes1a := runes1[:x]
	runes2a := runes2[:y]
	runes1b := runes1[x:]
	runes2b := runes2[y:]

	// Compute both diffsA serially.
	diffsA := diffMainRunes(runes1a, runes2a, deadline)
	diffsB := diffMainRunes(runes1b, runes2b, deadline)

	return append(diffsA, diffsB...)
}

func diffHalfMatch(text1, text2 []rune) [][]rune {
	var longText, shortText []rune
	if len(text1) > len(text2) {
		longText = text1
		shortText = text2
	} else {
		longText = text2
		shortText = text1
	}

	if len(longText) < 4 || len(shortText)*2 < len(longText) {
		return nil // Pointless.
	}

	// First check if the second quarter is the seed for a half-match.
	hm1 := diffHalfMatchI(longText, shortText, int(float64(len(longText)+3)/4)) // nolint: mnd

	// Check again based on the third quarter.
	hm2 := diffHalfMatchI(longText, shortText, int(float64(len(longText)+1)/2)) // nolint: mnd

	var halfMatch [][]rune

	switch {
	case hm1 == nil && hm2 == nil:
		return nil

	case hm2 == nil:
		halfMatch = hm1

	case hm1 == nil:
		halfMatch = hm2

	default:
		// Both matched. Select the longest.
		if len(hm1[4]) > len(hm2[4]) {
			halfMatch = hm1
		} else {
			halfMatch = hm2
		}
	}

	// A half-match was found, sort out the return data.
	if len(text1) > len(text2) {
		return halfMatch
	}

	return [][]rune{halfMatch[2], halfMatch[3], halfMatch[0], halfMatch[1], halfMatch[4]}
}

// diffHalfMatchI checks if a substring of shortText exist within longText such that the substring is at least half the length of longText.
// Returns a slice containing the prefix of longText, the suffix of longText, the prefix of shortText, the suffix of shortText and the common middle,
// or null if there was no match.
func diffHalfMatchI(l, s []rune, i int) [][]rune { // nolint: varnamelen
	var (
		bestCommonA    []rune
		bestCommonB    []rune
		bestCommonLen  int
		bestLongTextA  []rune
		bestLongTextB  []rune
		bestShortTextA []rune
		bestShortTextB []rune
	)

	// Start with a 1/4 length substring at position i as a seed.
	seed := l[i : i+len(l)/4]

	for j := runesIndexOf(s, seed, 0); j != -1; j = runesIndexOf(s, seed, j+1) { // nolint: varnamelen
		prefixLength := commonPrefixLength(l[i:], s[j:])
		suffixLength := commonSuffixLength(l[:i], s[:j])

		if bestCommonLen < suffixLength+prefixLength {
			bestCommonA = s[j-suffixLength : j]
			bestCommonB = s[j : j+prefixLength]
			bestCommonLen = len(bestCommonA) + len(bestCommonB)
			bestLongTextA = l[:i-suffixLength]
			bestLongTextB = l[i+prefixLength:]
			bestShortTextA = s[:j-suffixLength]
			bestShortTextB = s[j+prefixLength:]
		}
	}

	if bestCommonLen*2 < len(l) {
		return nil
	}

	return [][]rune{
		bestLongTextA,
		bestLongTextB,
		bestShortTextA,
		bestShortTextB,
		append(bestCommonA, bestCommonB...),
	}
}

// DiffCleanupMerge reorders and merges like edit sections. Merge equalities.
// Any edit section can move as long as it doesn't cross an equality.
// nolint: gocognit, cyclop, funlen
func DiffCleanupMerge(diffs []Diff) []Diff {
	// Add a dummy entry at the end.
	diffs = append(diffs, Diff{DiffEqual, ""})
	pointer := 0
	countDelete := 0
	countInsert := 0
	textDelete := []rune(nil)
	textInsert := []rune(nil)

	for pointer < len(diffs) {
		switch diffs[pointer].Type {
		case DiffInsert:
			countInsert++

			textInsert = append(textInsert, []rune(diffs[pointer].Text)...)
			pointer++

		case DiffDelete:
			countDelete++

			textDelete = append(textDelete, []rune(diffs[pointer].Text)...)
			pointer++

		case DiffEqual:
			// Upon reaching an equality, check for prior redundancies.
			switch {
			case countDelete+countInsert > 1:
				if countDelete != 0 && countInsert != 0 { //nolint: gocritic, nestif
					// Factor out any common prefixes.
					commonLength := commonPrefixLength(textInsert, textDelete)
					if commonLength != 0 {
						x := pointer - countDelete - countInsert
						if x > 0 && diffs[x-1].Type == DiffEqual {
							diffs[x-1].Text += string(textInsert[:commonLength])
						} else {
							diffs = append([]Diff{{DiffEqual, string(textInsert[:commonLength])}}, diffs...)
							pointer++
						}

						textInsert = textInsert[commonLength:]
						textDelete = textDelete[commonLength:]
					}
					// Factor out any common suffixes.
					commonLength = commonSuffixLength(textInsert, textDelete)
					if commonLength != 0 {
						insertIndex := len(textInsert) - commonLength
						deleteIndex := len(textDelete) - commonLength
						diffs[pointer].Text = string(textInsert[insertIndex:]) + diffs[pointer].Text
						textInsert = textInsert[:insertIndex]
						textDelete = textDelete[:deleteIndex]
					}
				}
				// Delete the offending records and add the merged ones.
				switch {
				case countDelete == 0:
					diffs = splice(diffs, pointer-countInsert,
						countDelete+countInsert,
						Diff{Type: DiffInsert, Text: string(textInsert)})

				case countInsert == 0:
					diffs = splice(diffs, pointer-countDelete,
						countDelete+countInsert,
						Diff{Type: DiffDelete, Text: string(textDelete)})

				default:
					diffs = splice(diffs, pointer-countDelete-countInsert,
						countDelete+countInsert,
						Diff{Type: DiffDelete, Text: string(textDelete)},
						Diff{Type: DiffInsert, Text: string(textInsert)})
				}

				pointer = pointer - countDelete - countInsert + 1
				if countDelete != 0 {
					pointer++
				}

				if countInsert != 0 {
					pointer++
				}

			case pointer != 0 && diffs[pointer-1].Type == DiffEqual:
				// Merge this equality with the previous one.
				diffs[pointer-1].Text += diffs[pointer].Text
				diffs = append(diffs[:pointer], diffs[pointer+1:]...)

			default:
				pointer++
			}

			countInsert = 0
			countDelete = 0
			textDelete = nil
			textInsert = nil
		}
	}

	if len(diffs[len(diffs)-1].Text) == 0 {
		diffs = diffs[0 : len(diffs)-1] // Remove the dummy entry at the end.
	}

	// Second pass: look for single edits surrounded on both sides by equalities which can be shifted sideways to eliminate an equality.
	// E.g: A<ins>BA</ins>C -> <ins>AB</ins>AC
	changes := false
	pointer = 1
	// Intentionally ignore the first and last element (don't need checking).
	for pointer < (len(diffs) - 1) {
		if diffs[pointer-1].Type == DiffEqual &&
			diffs[pointer+1].Type == DiffEqual {
			// This is a single edit surrounded by equalities.
			if strings.HasSuffix(diffs[pointer].Text, diffs[pointer-1].Text) {
				// Shift the edit over the previous equality.
				diffs[pointer].Text = diffs[pointer-1].Text +
					diffs[pointer].Text[:len(diffs[pointer].Text)-len(diffs[pointer-1].Text)]
				diffs[pointer+1].Text = diffs[pointer-1].Text + diffs[pointer+1].Text
				diffs = splice(diffs, pointer-1, 1)
				changes = true
			} else if strings.HasPrefix(diffs[pointer].Text, diffs[pointer+1].Text) {
				// Shift the edit over the next equality.
				diffs[pointer-1].Text += diffs[pointer+1].Text
				diffs[pointer].Text = diffs[pointer].Text[len(diffs[pointer+1].Text):] + diffs[pointer+1].Text
				diffs = splice(diffs, pointer+1, 1)
				changes = true
			}
		}

		pointer++
	}

	// If shifts were made, the diff needs reordering and another shift sweep.
	if changes {
		diffs = DiffCleanupMerge(diffs)
	}

	return diffs
}

// DiffCharsToLines rehydrates the text in a diff from a string of line hashes to real lines of text.
func DiffCharsToLines(diffs []Diff, lineArray []string) []Diff {
	hydrated := make([]Diff, 0, len(diffs))
	for _, aDiff := range diffs {
		runes := []rune(aDiff.Text)
		text := make([]string, len(runes))

		for i, r := range runes {
			text[i] = lineArray[runeToInt(r)]
		}

		aDiff.Text = strings.Join(text, "")
		hydrated = append(hydrated, aDiff)
	}

	return hydrated
}
