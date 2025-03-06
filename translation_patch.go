package divido

import (
	"io"
	"maps"
	"slices"
	"strings"
)

type xhtmlDecorations struct {
	prefix, suffix int
}

type translationPatch struct {
	sourceLines        []string
	contentDecorations map[int]xhtmlDecorations
	translatedContent  []string
}

func NewTranslationPatch(lines ...string) *translationPatch {
	return &translationPatch{
		sourceLines: lines,
	}
}

func (tp *translationPatch) UpdateContentDecorations() {

	tp.contentDecorations = make(map[int]xhtmlDecorations)

	for li, line := range tp.sourceLines {
		prefix, suffix := xhtmlDecorationsForLine(line)
		if prefix == -1 {
			continue
		}
		tp.contentDecorations[li] = xhtmlDecorations{prefix: prefix, suffix: suffix}
	}
}

func xhtmlDecorationsForLine(line string) (int, int) {
	prefix := strings.Index(line, ">")
	if prefix != -1 && prefix < len(line) {
		prefix += 1
	}
	suffix := strings.LastIndex(line, "<")
	if prefix > suffix {
		prefix = -1
	}
	return prefix, suffix
}

func (tp *translationPatch) SourceContent() []string {

	content := make([]string, 0, len(tp.contentDecorations))

	order := maps.Keys(tp.contentDecorations)
	sortedOrder := slices.Sorted(order)

	for _, li := range sortedOrder {
		xd := tp.contentDecorations[li]
		if xd.prefix == -1 {
			continue
		}
		content = append(content, tp.sourceLines[li][xd.prefix:xd.suffix])
	}

	return content
}

func (tp *translationPatch) AddTranslatedContent(tc []string) {
	tp.translatedContent = append(tp.translatedContent, tc...)
}

func (tp *translationPatch) Apply(w io.Writer) error {

	index := 0
	for li, line := range tp.sourceLines {
		if xd, ok := tp.contentDecorations[li]; ok {
			// redecorate
			translatedLine := tp.translatedContent[index]

			line = line[:xd.prefix] + translatedLine + line[xd.suffix:]

			index++
		}

		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
	}

	return nil
}
