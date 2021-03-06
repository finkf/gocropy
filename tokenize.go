package gocropy

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"unicode"
)

func sanitize(hocr *HOCR) error {
	pageBbox, err := hocr.ReadImageFileBbox()
	if err != nil {
		return err
	}
	for i := range hocr.Body.Div.Spans {
		bbox := hocr.Body.Div.Spans[i].GetBbox()
		bbox.Sanitize(*pageBbox)
		hocr.Body.Div.Spans[i].Title = fmt.Sprintf("bbox %v", bbox)
		hocr.Body.Div.Spans[i].Data = strings.Replace(
			hocr.Body.Div.Spans[i].Data, "\\&", "&", -1,
		)
		hocr.Body.Div.Spans[i].Data = strings.Replace(
			hocr.Body.Div.Spans[i].Data, "\\<", "<", -1,
		)
	}
	return nil
}

func appendCapability(metas []HOCRMeta) {
	for i := range metas {
		if metas[i].Name == "ocr-capabilities" {
			metas[i].Content = strings.Join(
				[]string{metas[i].Content, "ocrx_word"}, " ",
			)
		}
	}
}

type fileInfoByName struct {
	fs []os.FileInfo
}

func (f fileInfoByName) Len() int {
	return len(f.fs)
}

func (f fileInfoByName) Less(i, j int) bool {
	return f.fs[i].Name() < f.fs[j].Name()
}

func (f fileInfoByName) Swap(i, j int) {
	tmp := f.fs[i]
	f.fs[i] = f.fs[j]
	f.fs[j] = tmp
}

func readLlocs(dirname string) ([][]Lloc, error) {
	dir, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	sort.Sort(fileInfoByName{fileInfos})
	llocs := make([][]Lloc, 0, len(fileInfos)/3)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".llocs") {
			tmpLlocs, err := ReadLlocs(
				path.Join(dirname, fileInfo.Name()),
			)
			if err != nil {
				return nil, err
			}
			llocs = append(llocs, tmpLlocs)
		}
	}
	return llocs, nil
}

func tokenizeSpan(llocs []Lloc, span *HOCRSpan) {
	chars := CharsFromLlocs(llocs, span.GetBbox())
	// append one trailing whitespace in order to add the last token
	chars = append(chars, Char{Bbox{0, 0, 0, 0}, ' '})

	n := 0
	for i := range chars {
		if unicode.IsSpace(chars[i].Rune) && n > 0 {
			token := chars[(i - n):i]
			str, bbox := TokenizeChars(token)
			tspan := HOCRSpan{Data: str}
			tspan.Class = "ocrx_word"
			tspan.SetBbox(bbox)
			tspan.SetCuts(token)
			span.Token = append(span.Token, tspan)
			n = 0
		} else if !unicode.IsSpace(chars[i].Rune) {
			n++
		}
	}
	// remove line data
	span.Data = ""
}

func tokenize(hocr *HOCR, dir string) error {
	llocs, err := readLlocs(dir)
	if err != nil {
		return err
	}
	if len(llocs) != len(hocr.Body.Div.Spans) {
		return fmt.Errorf(
			"Number of lines in HOCR (%v) differ from number of llocs (%v) in `%v`",
			len(hocr.Body.Div.Spans),
			len(llocs),
			dir,
		)
	}
	for i := range llocs {
		tokenizeSpan(llocs[i], &hocr.Body.Div.Spans[i])
	}
	appendCapability(hocr.Head.Metas)
	return nil
}

func (hocr *HOCR) ConvertToHOCR(dir string) error {
	err := sanitize(hocr)
	if err != nil {
		return err
	}
	return tokenize(hocr, dir)
}

func (hocr *HOCR) MustConvertToHOCR(dir string) {
	if err := hocr.ConvertToHOCR(dir); err != nil {
		panic(err)
	}
}
