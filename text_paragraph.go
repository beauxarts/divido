package divido

import "strings"

type TextParagraph string

func NewParagraphs(p string) []TextParagraph {
	return []TextParagraph{TextParagraph(p)}
}

func (tp TextParagraph) String() string {
	return string(tp)
}

func (tp TextParagraph) Sentences() []string {
	return strings.Split(string(tp), ". ")
}
