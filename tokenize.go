package gocropy

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"unicode"
)

func sanitize(hocr *Hocr) error {
	pageBbox, err := hocr.ReadImageFileBbox()
	if err != nil {
		return err
	}
	for i := range hocr.Body.Div.Spans {
		bbox := hocr.Body.Div.Spans[i].GetBbox()
		bbox.Sanitize(*pageBbox)
		hocr.Body.Div.Spans[i].Title = fmt.Sprintf("%v", bbox)
		hocr.Body.Div.Spans[i].Data = strings.Replace(
			hocr.Body.Div.Spans[i].Data, "\\&", "&", -1,
		)
		hocr.Body.Div.Spans[i].Data = strings.Replace(
			hocr.Body.Div.Spans[i].Data, "\\<", "<", -1,
		)
	}
	return nil
}

func appendCapability(metas []HocrMeta) {
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

type char struct {
	r    rune
	bbox Bbox
}

func makeChar(span HocrSpan, lloc Lloc) char {
	spanBbox := span.GetBbox()
	return char{
		lloc.Codepoint,
		Bbox{
			spanBbox.Left + int64(lloc.Adjustment),
			spanBbox.Top,
			0,
			spanBbox.Bottom,
		},
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func combineBbox(chars []char, i, n int) Bbox {
	bbox := chars[i].bbox
	for i := i + 1; i < n; i++ {
		bbox.Left = min(bbox.Left, chars[i].bbox.Left)
		bbox.Top = min(bbox.Top, chars[i].bbox.Top)
		bbox.Right = max(bbox.Right, chars[i].bbox.Right)
		bbox.Bottom = max(bbox.Bottom, chars[i].bbox.Bottom)
	}
	return bbox
}

func tokenizeSpan(llocs []Lloc, span *HocrSpan) {
	chars := make([]char, 0, len(llocs)+1)
	for i := range llocs {
		chars = append(chars, makeChar(*span, llocs[i]))
		if i > 0 {
			chars[i-1].bbox.Right = chars[i].bbox.Left - 1
		}
	}
	// append one trailing whitespace in order to add the last token
	chars = append(chars, char{' ', Bbox{0, 0, 0, 0}})
	var buffer bytes.Buffer
	n := 0
	for i := range chars {
		if unicode.IsSpace(chars[i].r) && n > 0 {
			var tspan HocrSpan
			tspan.Class = "ocrx_word"
			tspan.SetBbox(combineBbox(chars, i-n, n))
			tspan.Data = buffer.String()
			buffer.Reset()
			n = 0
			span.Token = append(span.Token, tspan)
		} else if !unicode.IsSpace(chars[i].r) {
			buffer.WriteRune(chars[i].r)
			n++
		}
	}
	// remove line data
	span.Data = ""
}

func tokenize(hocr *Hocr, dir string) error {
	llocs, err := readLlocs(dir)
	if err != nil {
		return err
	}
	if len(llocs) != len(hocr.Body.Div.Spans) {
		return fmt.Errorf(
			"Number of lines in Hocr (%v) differ from number of llocs (%v) in `%v`",
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

func (hocr *Hocr) ConvertToHocr(dir string) error {
	err := sanitize(hocr)
	if err == nil {
		return tokenize(hocr, dir)
	} else {
		return err
	}
}

func (hocr *Hocr) MustConvertToHocr(dir string) {
	if err := hocr.ConvertToHocr(dir); err != nil {
		panic(err)
	}
}
