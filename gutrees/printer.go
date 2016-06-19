package gutrees

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-humble/detect"
)

//==============================================================================

// Mode defines the behaviour of the printer setup. Where each
// increasing mode affects the behaviour and final form of the
// print outs.
type Mode int

const (
	// Normal mode means all Ids, Hashes are printed and
	// Removals are cleaned out.
	Normal Mode = iota

	// Testing mode means all Ids and Hashes are not printed
	// but all removals are cleaned out to allow a predictable output.
	Testing

	// Debugging mode means all Ids and Hashes are not printed and
	// all removals are left behind to ensure deugging is possible.
	// This allows us to see the state of a reconciled tree.
	Debugging
)

// currentMode defines the struct which manages the
// mode of operation for the printer.
var cu = struct {
	r sync.Mutex
	m Mode
}{
	m: Normal,
}

// GetMode returns the current working mode for the libraries printers.
func GetMode() Mode {
	cu.r.Lock()
	defer cu.r.Unlock()
	return cu.m
}

// SetMode sets the working mode for the library printers.
func SetMode(ms Mode) {
	cu.r.Lock()
	defer cu.r.Unlock()
	cu.m = ms
}

//==============================================================================

// AttrPrinter defines a printer interface for writing out a Attribute objects into a string form
type AttrPrinter interface {
	Print([]Property) string
}

// SimpleAttrWriter provides a basic attribute writer
var SimpleAttrWriter AttrWriter

// AttrWriter provides a concrete struct that meets the AttrPrinter interface
type AttrWriter struct{}

const attrformt = ` %s="%s"`

// Print returns a stringed repesentation of the attribute object
func (m AttrWriter) Print(a []Property) string {
	if len(a) <= 0 {
		return ""
	}

	attrs := []string{}

	for _, ar := range a {
		name, val := ar.Render()
		attrs = append(attrs, fmt.Sprintf(attrformt, name, val))
	}

	return strings.Join(attrs, " ")
}

//==============================================================================

// StylePrinter defines a printer interface for writing out a style objects into a string form
type StylePrinter interface {
	Print([]Property) string
}

// SimpleStyleWriter provides a basic style writer
var SimpleStyleWriter StyleWriter

// StyleWriter provides a concrete struct that meets the AttrPrinter interface
type StyleWriter struct{}

const styleformt = " %s:%s;"

// Print returns a stringed repesentation of the style object
func (m StyleWriter) Print(s []Property) string {
	if len(s) <= 0 {
		return ""
	}

	css := []string{}

	for _, cs := range s {
		name, val := cs.Render()
		css = append(css, fmt.Sprintf(styleformt, name, val))
	}

	return strings.Join(css, " ")
}

//==============================================================================

// TextPrinter defines a printer interface for writing out a text type markup into a string form
type TextPrinter interface {
	Print(Markup) string
}

// TextWriter writes out the text element/node for the vdom into a string
type TextWriter struct{}

// SimpleTextWriter provides a basic text writer
var SimpleTextWriter TextWriter

// Print returns the string representation of the text object
func (m TextWriter) Print(t Markup) string {
	if tt, ok := t.(TextMarkup); ok {
		return tt.TextContent()
	}

	return ""
}

//==============================================================================

// MarkupWriter defines a printer interface for writing out a markup object into a string form
type MarkupWriter interface {
	Write(Markup) (string, error)
}

// ElementWriter writes out the element out as a string matching the html tag rules
type ElementWriter struct {
	attrWriter  AttrPrinter
	styleWriter StylePrinter
	text        TextPrinter
}

// SimpleElementWriter provides a default writer using the basic attribute and style writers
var SimpleElementWriter = NewElementWriter(SimpleAttrWriter, SimpleStyleWriter, SimpleTextWriter)

// NewElementWriter returns a new writer for Element objects
func NewElementWriter(aw AttrPrinter, sw StylePrinter, tw TextPrinter) *ElementWriter {
	return &ElementWriter{
		attrWriter:  aw,
		styleWriter: sw,
		text:        tw,
	}
}

// Write prints the giving Markup as a string else returns an error.
func (m *ElementWriter) Write(ma Markup) (string, error) {
	if emr, ok := ma.(*Element); ok {
		return m.Print(emr), nil
	}

	return "", ErrNotMarkup
}

// Print returns the string representation of the element
func (m *ElementWriter) Print(e *Element) string {

	// if we are on the server && is this element marked as removed, if so we skip and return an empty string
	if detect.IsServer() {
		if e.Removed() && GetMode() < Debugging {
			return ""
		}
	}

	//if we are dealing with a text type just return the content
	if e.Name() == "text" {
		return m.text.Print(e)
	}

	//management attributes
	var mido []Property

	//collect uid and hash of the element so we can write them along
	if GetMode() < Testing {
		hash := &Attribute{"hash", e.Hash()}
		uid := &Attribute{"uid", e.UID()}
		mido = append(mido, hash, uid)
	}

	// if e.Removed() && GetMode() > Testing {
	// 	mido = append(mido, Attribute{"RemovedNode", "true"})
	// }

	//write out the hash and uid as attributes
	hashes := m.attrWriter.Print(mido)

	//write out the elements attributes using the AttrWriter
	attrs := m.attrWriter.Print(e.Attributes())

	//write out the elements inline-styles using the StyleWriter
	style := m.styleWriter.Print(e.Styles())

	var closer string
	var beginbrack string

	if e.AutoClosed() {
		closer = "/>"
	} else {
		beginbrack = ">"
		closer = fmt.Sprintf("</%s>", e.Name())
	}

	var children = []string{}

	for _, ch := range e.Children() {
		if ech, ok := ch.(*Element); ok {
			if ech == e {
				continue
			}
			children = append(children, m.Print(ech))
		}
	}

	//lets create the elements markup now
	return strings.Join([]string{
		fmt.Sprintf("<%s", e.Name()),
		hashes,
		attrs,
		fmt.Sprintf(` style="%s"`, style),
		beginbrack,
		e.textContent,
		strings.Join(children, ""),
		closer,
	}, "")
}

//==============================================================================
