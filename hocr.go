package gocropy

import (
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

type HocrTitle struct {
	Title string `xml:"title,attr"`
}

type HocrClass struct {
	Class string `xml:"class,attr"`
}

type HocrSpan struct {
	HocrClass
	HocrTitle
	Data  string     `xml:",chardata"`
	Token []HocrSpan `xml:"span"`
}

type HocrDiv struct {
	HocrClass
	HocrTitle
	Spans []HocrSpan `xml:"span"`
}

type HocrBody struct {
	Div HocrDiv `xml:"div"`
}

type HocrMeta struct {
	HttpEquiv string `xml:"http-equiv,attr,omitempty"`
	Name      string `xml:"name,attr,omitempty"`
	Content   string `xml:"content,attr"`
}

type HocrHead struct {
	Title string     `xml:"title"`
	Metas []HocrMeta `xml:"meta"`
}

type Hocr struct {
	XMLName xml.Name
	Head    HocrHead `xml:"head"`
	Body    HocrBody `xml:"body"`
	Path    string
}

func (div *HocrDiv) MustReadImageFileBbox() Bbox {
	in, err := os.Open(div.GetFile())
	if err == nil {
		defer in.Close()
		im, _, err := image.DecodeConfig(in)
		if err == nil {
			if im.ColorModel == nil {
				panic("Could not read image (missing decoder?)")
			}
			bbox := Bbox{0, 0, int64(im.Width), int64(im.Height)}
			div.SetBbox(bbox)
			return bbox
		}
	}
	panic(err)
}

var fileRegex = regexp.MustCompile("file\\s+(.*)")

func (title HocrTitle) GetFile() string {
	m := fileRegex.FindStringSubmatch(title.Title)
	if m != nil {
		return m[1]
	} else {
		return ""
	}
}

func (title *HocrTitle) SetFile(file string) {
	if len(title.Title) > 0 {
		title.Title = fmt.Sprintf("%s; file %s", title.Title, file)
	} else {
		title.Title = fmt.Sprintf("file %s", file)
	}
}

var bboxRegex = regexp.MustCompile("bbox\\s+(\\d+)\\s+(\\d+)\\s+(\\d+)\\s+(\\d+)")

func (title HocrTitle) GetBbox() Bbox {
	var bbox = Bbox{-1, -1, -1, -1}
	m := bboxRegex.FindStringSubmatch(title.Title)
	if m != nil {
		bbox.Left, _ = strconv.ParseInt(m[1], 0, 64)
		bbox.Top, _ = strconv.ParseInt(m[2], 0, 64)
		bbox.Right, _ = strconv.ParseInt(m[3], 0, 64)
		bbox.Bottom, _ = strconv.ParseInt(m[4], 0, 64)
	}
	return bbox
}

func (title *HocrTitle) SetBbox(bbox Bbox) {
	if len(title.Title) > 0 {
		title.Title = fmt.Sprintf("%s; %v", title.Title, bbox)
	} else {
		title.Title = fmt.Sprintf("%v", bbox)
	}
}

func ReadHocr(file string) (*Hocr, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	hocr := Hocr{Path: file}
	err = xml.Unmarshal(data, &hocr)
	if err != nil {
		return nil, err
	}
	return &hocr, nil
}

func MustReadHocr(file string) *Hocr {
	hocr, err := ReadHocr(file)
	if err != nil {
		panic(err)
	}
	return hocr
}
