package main

import (
	"fmt"
	"os"
	"strconv"
	"word-freq/analysis"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("The first argument should have the filename and the second should be n")
		return
	}
	path := os.Args[1]
	n, err := strconv.Atoi(string(os.Args[2]))
	if err != nil {
		panic(err)
	}
	text, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	topWords := analysis.TopN(string(text), n)
	for _, wordf := range topWords {
		fmt.Printf("%v: %d\n", wordf.Word, wordf.Count)
	}
}
