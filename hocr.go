package gocropy

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	//	"strconv"
)

type HOCRTitle struct {
	Title string `xml:"title,attr"`
}

type HOCRClass struct {
	Class string `xml:"class,attr"`
}

type HOCRSpan struct {
	HOCRClass
	HOCRTitle
	Data  string     `xml:",chardata"`
	Token []HOCRSpan `xml:"span"`
}

type HOCRDiv struct {
	HOCRClass
	HOCRTitle
	Spans []HOCRSpan `xml:"span"`
}

type HOCRBody struct {
	Div HOCRDiv `xml:"div"`
}

type HOCRMeta struct {
	HttpEquiv string `xml:"http-equiv,attr,omitempty"`
	Name      string `xml:"name,attr,omitempty"`
	Content   string `xml:"content,attr"`
}

type HOCRHead struct {
	Title string     `xml:"title"`
	Metas []HOCRMeta `xml:"meta"`
}

type HOCR struct {
	XMLName xml.Name
	Head    HOCRHead `xml:"head"`
	Body    HOCRBody `xml:"body"`
	File    string   `xml:"-"`
}

func (hocr *HOCR) ReadImageFileBbox() (*Bbox, error) {
	basedir, _ := path.Split(hocr.File)
	file := hocr.Body.Div.GetFile()
	in, err := os.Open(path.Join(basedir, file))
	if err != nil {
		return nil, err
	}
	defer in.Close()
	img, _, err := image.DecodeConfig(in)
	if err != nil {
		return nil, err
	}
	if img.ColorModel == nil {
		return nil, fmt.Errorf("Could not read image (missing decoder?)")
	}
	bbox := Bbox{0, 0, img.Width, img.Height}
	hocr.Body.Div.SetBbox(bbox)
	return &bbox, nil
}

var fileRegex = regexp.MustCompile("file\\s+(.*)")

func (title HOCRTitle) GetFile() string {
	m := fileRegex.FindStringSubmatch(title.Title)
	if m != nil {
		return m[1]
	} else {
		return ""
	}
}

func (title *HOCRTitle) SetFile(file string) {
	if len(title.Title) > 0 {
		title.Title = fmt.Sprintf("%s; file %s", title.Title, file)
	} else {
		title.Title = fmt.Sprintf("file %s", file)
	}
}

func (title *HOCRTitle) SetCuts(chars []Char) {
	var buffer bytes.Buffer
	for i := 1; i < len(chars); i++ {
		delta := chars[i].Box.Left - chars[i-1].Box.Left
		buffer.WriteString(fmt.Sprintf(" %d", delta))
	}
	if len(title.Title) > 0 {
		title.Title = fmt.Sprintf("%s; cuts%s", title.Title, buffer.String())
	} else {
		title.Title = fmt.Sprintf("cuts%s", buffer.String())
	}
}

var bboxRegex = regexp.MustCompile("bbox\\s+(\\d+\\s+\\d+\\s+\\d+\\s+\\d+)")

func (title HOCRTitle) GetBbox() Bbox {
	var bbox Bbox
	m := bboxRegex.FindStringSubmatch(title.Title)
	if m != nil {
		fmt.Sscanf(m[1], "%v", &bbox)
	}
	return bbox
}

func (title *HOCRTitle) SetBbox(bbox Bbox) {
	if len(title.Title) > 0 {
		title.Title = fmt.Sprintf("%s; bbox %v", title.Title, bbox)
	} else {
		title.Title = fmt.Sprintf("bbox %v", bbox)
	}
}

func ReadHOCR(file string) (*HOCR, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	hocr := HOCR{File: file}
	err = xml.Unmarshal(data, &hocr)
	if err != nil {
		return nil, err
	}
	return &hocr, nil
}

func MustReadHOCR(file string) *HOCR {
	hocr, err := ReadHOCR(file)
	if err != nil {
		panic(err)
	}
	return hocr
}
