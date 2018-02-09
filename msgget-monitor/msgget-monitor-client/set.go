package main

import (
	"strings"
)

type StringSet struct {
	S map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		S: make(map[string]struct{}),
	}
}

func (this *StringSet) Add(element string) *StringSet {
	this.S[element] = struct{}{}
	return this
}

func (this *StringSet) Exists(element string) bool {
	_, exists := this.S[element]
	return exists
}

func (this *StringSet) Delete(element string) {
	delete(this.S, element)
}

func (this *StringSet) Clear() {
	this.S = make(map[string]struct{})
}

func (this *StringSet) Len() int {
	return len(this.S)
}

func (this *StringSet) ToValueString() string {
	count := len(this.S)
	if count == 0 {
		return ""
	}

	result := ""
	for element := range this.S {
		result = result + " " + element
	}
	return strings.Trim(result, " ")
}

func (this *StringSet) ToSlice() []string {
	count := len(this.S)
	if count == 0 {
		return []string{}
	}

	r := make([]string, count)

	i := 0
	for element := range this.S {
		r[i] = element
		i++
	}

	return r
}
