package divido

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	significantBreaksInRow = 3
	notesSuffix            = "_notes"
)

type TextDocument []TextElement

func (td TextDocument) ChapterTitles() []string {
	chapters := make([]string, 0, len(td)/2)
	for _, tc := range td {
		if tc.Type == ChapterTitle {
			for _, p := range tc.Content {
				chapters = append(chapters, p.String())
			}
		}
	}
	return chapters
}

func (td TextDocument) ChapterParagraphs(chapterTitle string) []TextParagraph {
	paragraphs := make([]TextParagraph, 0)
	accumulate := false
	for _, tc := range td {
		if tc.Type == Break {
			continue
		}
		if tc.Type == ChapterTitle {
			if accumulate == true {
				break
			} else {
				if len(tc.Content) > 0 && tc.Content[0].String() == chapterTitle {
					accumulate = true
				}
			}
		}
		if tc.Type == Paragraph && accumulate {
			paragraphs = append(paragraphs, tc.Content...)
		}
	}
	return paragraphs
}

func (td TextDocument) ReplaceFrom(start int, old, new string) int {

	if start >= len(td) {
		return -1
	}

	for ii := start; ii < len(td); ii++ {
		n := td[ii]
		if n.Type != Paragraph {
			continue
		}
		for pi, p := range n.Content {
			if strings.Contains(p.String(), old) {
				n.Content[pi] = TextParagraph(strings.Replace(p.String(), old, new, 1))
				return ii
			}
		}
	}

	return -1
}

func NewTextDocument(reader io.Reader) TextDocument {
	//divido includes the following steps to process plain text into structured text:
	//1) determine significant breaks (>1 \n character in a row), paragraphs
	//2) mark chapter titles and in-chapter dividers

	scanner := bufio.NewScanner(reader)

	td := make(TextDocument, 0)
	breaksInRow := 0

	// 1)
	for scanner.Scan() {

		paragraphCandidate := scanner.Text()

		if len(paragraphCandidate) == 0 {
			breaksInRow++
		} else {
			if breaksInRow > significantBreaksInRow {
				td = append(td, TextElement{
					Type: Break,
				})
			}
			if len(td) > 0 &&
				td[len(td)-1].Type == Paragraph {
				td[len(td)-1].Content = append(td[len(td)-1].Content, TextParagraph(paragraphCandidate))
			} else {
				td = append(td, TextElement{
					Content: NewParagraphs(paragraphCandidate),
					Type:    Paragraph,
				})
			}

			breaksInRow = 0
		}
	}

	//2)

	for ii, tc := range td {
		if tc.Type == Paragraph && len(tc.Content) == 1 {
			if ii < len(td)-1 && td[ii+1].Type == Break {
				td[ii].Type = ChapterTitle
			}
		}
	}

	return td
}

func NewTextDocumentWithNotes(document io.Reader, notes io.Reader) TextDocument {

	td := NewTextDocument(document)
	nt := NewTextDocument(notes)

	previousType := Break

	// reformat nodes to have ChapterTitle, then Paragraph pattern
	for ii, tc := range nt {
		if tc.Type == ChapterTitle {
			if previousType == ChapterTitle {
				nt[ii].Type = Paragraph
				previousType = Paragraph
			} else {
				previousType = ChapterTitle
			}
		}
	}

	noteTitles := nt.ChapterTitles()
	if len(noteTitles) == 0 {
		return td
	}

	lastIndex := 0

	for _, ni := range nt.ChapterTitles() {
		noteIndex := fmt.Sprintf("[%s]", ni)
		for _, nc := range nt.ChapterParagraphs(ni) {
			noteText := fmt.Sprintf(" (%s)", nc.String())
			if lastIndex = td.ReplaceFrom(lastIndex, noteIndex, noteText); lastIndex < 0 {
				break
			}
		}
	}

	return td
}

func DefaultNotesFilename(filename string) string {
	ext := filepath.Ext(filename)
	filenameSansExt := strings.TrimSuffix(filename, ext)
	return filenameSansExt + notesSuffix + ext
}
