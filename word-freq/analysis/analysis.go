package analysis

import (
	"slices"
	"sort"
	"strings"
)

type WordFrequency struct {
	Word  string
	Count int
}

// returns -1 if not in word list, otherwise return the index
func findWordFrequencyIndex(word string, topNWords []WordFrequency) int {
	return slices.IndexFunc(topNWords, func(wf WordFrequency) bool {
		return wf.Word == word
	})
}

// return a list of word frequency, limited by n
func TopN(text string, n int) []WordFrequency {
	split_string := strings.Fields(text)
	topNWords := make([]WordFrequency, 0)

	for _, w := range split_string {
		w = strings.ToLower(w)
		index := findWordFrequencyIndex(w, topNWords)
		if index < 0 {
			topNWords = append(topNWords, WordFrequency{w, 1})
		} else {
			topNWords[index].Count++
		}
	}

	sort.Slice(topNWords, func(i, j int) bool { return topNWords[i].Count >= topNWords[j].Count })

	if len(topNWords) > n {
		topNWords = topNWords[:n]
	}

	return topNWords
}
