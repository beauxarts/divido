package divido

import (
	"bufio"
	"io"
	"strings"
)

const (
	significantBreaksInRow = 3
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

func (td TextDocument) ExportMetadata(title, author string) string {
	sb := strings.Builder{}

	//https://ffmpeg.org/ffmpeg-all.html#Metadata-1
	sb.WriteString(";FFMETADATA1")
	if title != "" {
		sb.WriteString("title=" + title)
	}
	if author != "" {
		sb.WriteString("artist=" + author)
	}
	sb.WriteString("\n")

	for _, ct := range td.ChapterTitles() {
		sb.WriteString("[CHAPTER]")
		sb.WriteString("title=" + ct)
	}

	return sb.String()
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
