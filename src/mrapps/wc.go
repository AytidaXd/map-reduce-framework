package main

//
// a word-count application "plugin" for MapReduce.
//
// go build -buildmode=plugin wc.go
//

import "MapReduce/mr"
import "unicode"
import "strings"
import "strconv"

// Called once for each file of input. 
// Takes input filename and contents.
func Map(filename string, contents string) []mr.KeyValue {
	// function to detect word separators.
	ff := func(r rune) bool { return !unicode.IsLetter(r) }

	// split contents into an array of words.
	words := strings.FieldsFunc(contents, ff)

	kva := []mr.KeyValue{}
	for _, w := range words {
		kv := mr.KeyValue{w, "1"}
		kva = append(kva, kv)
	}
	return kva
}

// return the number of occurrences of this word.
func Reduce(key string, values []string) string {
	return strconv.Itoa(len(values))
}
