package parsers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

var (
	// errTagSyntax is an error for when a tag is not valid
	errTagSyntax = errors.New("bad syntax for struct tag pair")
	// errTagKeySyntax is an error for when a tag key is not valid
	errTagKeySyntax = errors.New("bad syntax for struct tag key")
	// errTagValueSyntax is an error for when a tag value is not valid
	errTagValueSyntax = errors.New("bad syntax for struct tag value")
	// errKeyNotSet is an error for when a tag key is not set
	errKeyNotSet = errors.New("tag key does not exist")
	// errTagNotExist is an error for when a tag does not exist
	errTagNotExist = errors.New("tag does not exist")
)

// Structure is a struct for a struct golang definition
type Structure struct {
	// Fields is a map of the fields in the struct
	Fields []Field `json:"fields"`
}

// Field is a struct for a field within a struct
type Field struct {
	// Name is the name of the field
	Name string `json:"name"`
	// Type is the type of the field actually in the struct
	Type string `json:"type"`
	// Line is the line of the field in the source code
	Line int `json:"line"`
	// Tags are the tags for the field
	Tags Tags `json:"tags"`
	// Start is the start of the tag in the source code
	Start int `json:"start"`
	// End is the end of the tag in the source code
	End int `json:"end"`
}

// Tags represent a set of tags from a single struct field
type Tags struct {
	tags []*Tag
}

// Tag defines a single struct's string literal tag
type Tag struct {
	// Key is the tag key, such as json, xml, etc..
	// i.e: `json:"foo,omitempty". Here key is: "json"
	Key string `json:"key"`
	// Name is a part of the value
	// i.e: `json:"foo,omitempty". Here name is: "foo"
	Name string `json:"name"`
	// Options is a part of the value. It contains a slice of tag options i.e:
	// `json:"foo,omitempty". Here options is: ["omitempty"]
	Options []string `json:"options"`
	// Line is the line of the tag in the source code
	Line int `json:"line"`
	// Start is the start of the tag in the source code horizontally
	Start int `json:"start"`
	// End is the end of the tag in the source code horizontally
	End int `json:"end"`
}

// Inspector is a function for concurrently inspecting a node recursively
type Inspector func(n ast.Node) bool

// ParseStructs checks if the struct tag is in the url and returns a
func ParseStructs(ctx context.Context, src []byte) (structures []Structure, err error) {
	// add a package main to the source
	strSrc := string(src)
	if !strings.Contains(strSrc, "package ") {
		strSrc = "package main\n" + strSrc
	}
	src = []byte(strSrc)
	var eg *errgroup.Group
	eg, _ = errgroup.WithContext(ctx)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct: %w", err)
	}
	ast.Inspect(file, func() func(n ast.Node) bool {
		return func(n ast.Node) bool {
			s, ok := n.(*ast.StructType)
			if !ok {
				return true
			}
			var structure Structure
			structure.Fields = make([]Field, len(s.Fields.List))
			for idx, field := range s.Fields.List {
				field, idx := field, idx
				eg.Go(func() error {
					tags, err := ParseTags(
						field.Tag.Value[1:len(field.Tag.Value)-1],
						fset.Position(field.Pos()).Offset,
						fset.Position(field.End()).Offset,
						fset.Position(field.Pos()).Line,
					)
					if err != nil {
						return fmt.Errorf("failed to parse tags: %w", err)
					}
					structure.Fields[idx] = Field{
						Name:  field.Names[0].Name,
						Type:  fmt.Sprintf("%s", field.Type),
						Tags:  *tags,
						Line:  fset.Position(field.Pos()).Line,
						Start: fset.Position(field.Pos()).Offset,
					}
					return nil
				})
			}
			if err := eg.Wait(); err != nil {
				return true
			}
			structures = append(structures, structure)
			return true
		}
	}())
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to parse struct: %w", err)
	}
	return structures, nil
}

// ParseTags parses a single struct field tag and returns the set of tags.
func ParseTags(tag string, start, end, line int) (*Tags, error) {
	var tags []*Tag
	hasTag := tag != ""
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		// Scan to colon. A space, a quote or a control character is a syntax
		// error. Strictly speaking, control chars include the range [0x7f,
		// 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
		// multi-byte control characters as it is simpler to inspect the tag's
		// bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 {
			return nil, errTagKeySyntax
		}
		if i+1 >= len(tag) || tag[i] != ':' {
			return nil, errTagSyntax
		}
		if tag[i+1] != '"' {
			return nil, errTagValueSyntax
		}
		key := tag[:i]
		tag = tag[i+1:]
		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, errTagValueSyntax
		}
		qvalue := tag[:i+1]
		tag = tag[i+1:]
		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, errTagValueSyntax
		}
		res := strings.Split(value, ",")
		name := res[0]
		options := res[1:]
		if len(options) == 0 {
			options = nil
		}
		tags = append(tags, &Tag{
			Key:     key,
			Name:    name,
			Options: options,
			Start:   start,
			End:     end,
			Line:    line,
		})
	}
	if hasTag && len(tags) == 0 {
		return nil, nil
	}
	return &Tags{
		tags: tags,
	}, nil
}

// Get returns the tag associated with the given key. If the key is present
// in the tag the value (which may be empty) is returned. Otherwise, the
// returned value will be the empty string. The ok return value reports whether
// the tag exists or not (which the return value is nil).
func (t *Tags) Get(key string) (*Tag, error) {
	for _, tag := range t.tags {
		if tag.Key == key {
			return tag, nil
		}
	}
	return nil, errTagNotExist
}

// Set sets the given tag. If the tag key already exists it'll override it
func (t *Tags) Set(tag *Tag) error {
	if tag.Key == "" {
		return errKeyNotSet
	}
	added := false
	for i, tg := range t.tags {
		if tg.Key == tag.Key {
			added = true
			t.tags[i] = tag
		}
	}
	if !added {
		// this means this is a new tag, add it
		t.tags = append(t.tags, tag)
	}
	return nil
}

// AddOptions adds the given option for the given key. If the option already
// exists it doesn't add it again.
func (t *Tags) AddOptions(key string, options ...string) {
	for i, tag := range t.tags {
		if tag.Key != key {
			continue
		}
		for _, opt := range options {
			if !tag.HasOption(opt) {
				tag.Options = append(tag.Options, opt)
			}
		}
		t.tags[i] = tag
	}
}

// Tags returns a slice of tags. The order is the original tag order unless it
// was changed.
func (t *Tags) Tags() []*Tag {
	return t.tags
}

// Keys returns a slice of tags' keys.
func (t *Tags) Keys() []string {
	var keys []string
	for _, tag := range t.tags {
		keys = append(keys, tag.Key)
	}
	return keys
}

// String reassembles the tags into a valid literal tag field representation
func (t *Tags) String() string {
	tags := t.Tags()
	if len(tags) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i, tag := range t.Tags() {
		buf.WriteString(tag.String())
		if i != len(tags)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

// HasOption returns true if the given option is available in options
func (t *Tag) HasOption(opt string) bool {
	for _, tagOpt := range t.Options {
		if tagOpt == opt {
			return true
		}
	}
	return false
}

// Value returns the raw value of the tag, i.e. if the tag is
// `json:"foo,omitempty", the Value is "foo,omitempty"
func (t *Tag) Value() string {
	options := strings.Join(t.Options, ",")
	if options != "" {
		return fmt.Sprintf(`%s,%s`, t.Name, options)
	}
	return t.Name
}

// GoString implements the fmt.GoStringer interface
func (t *Tag) GoString() string {
	template := `{
		Key:    '%s',
		Name:   '%s',
		Option: '%s',
	}`
	if t.Options == nil {
		return fmt.Sprintf(template, t.Key, t.Name, "nil")
	}
	options := strings.Join(t.Options, ",")
	return fmt.Sprintf(template, t.Key, t.Name, options)
}

// String returns a string representation of the tag
func (t *Tag) String() string {
	return fmt.Sprintf(`%s:%q`, t.Key, t.Value())
}
