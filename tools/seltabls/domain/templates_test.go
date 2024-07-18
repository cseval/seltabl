package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/conneroisu/seltabl/tools/seltabls/data/master"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// TestNewStructFileContent tests the NewStructFileContent struct
func TestNewStructFileContent(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		sectionErrorArgs{
			Error: fmt.Errorf(
				"failed to parse struct: failed to get data rows html: failed to get html: failed to get doc: open /Users/hsz/Projects/github.com/conneroisu/seltabl/testdata/ab_num_table.html: no such file or directory",
			),
			History: []anthropic.Message{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: []anthropic.MessageContent{anthropic.NewTextMessageContent("foo")},
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: []anthropic.MessageContent{anthropic.NewTextMessageContent("bar")},
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestNewAggregatePrompt tests the NewAggregatePrompt struct
func TestNewAggregatePrompt(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		SectionAggregateArgs{
			Structs: []string{"ex json 1 ", "ex json 2 ", "ex json 3 "},
			Content: `<html><body><table><tr><td>a</td><td>b</td></tr><tr><td>1</td><td>2</td></tr></table></body></html>`,
			Selectors: []master.Selector{
				{
					ID:         2,
					Value:      "dsaf",
					UrlID:      2,
					Occurances: 2,
					Context:    "<html>",
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

func TestIdentifyAggregateArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		IdentifyAggregateArgs{
			Content: "<html><body><table><tr><td>a</td><td>b</td></tr><tr><td>1</td><td>2</td></tr></table></body></html>",
			Schemas: []string{"dsafsd", "dsazfdasdfasf"},
			Selectors: []master.Selector{
				{
					ID:         2,
					Value:      "dsaf",
					UrlID:      2,
					Occurances: 2,
					Context:    "<html>",
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestStructPromptArgs tests the StructPromptArgs struct.
func TestStructPromptArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		StructPromptArgs{
			URL:     "https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			Content: "<html><body><table><tr><td>a</td><td>b</td></tr><tr><td>1</td><td>2</td></tr></table></body></html>",
			Selectors: []master.Selector{
				{
					ID:         1,
					Value:      "html > body > table > tbody > tr > td:nth-child(1)",
					Occurances: 1,
					Context:    "<html></html>",
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestStructFilePromptArgs tests the StructFilePromptArgs struct.
func TestStructFilePromptArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		StructFilePromptArgs{
			PackageName: "main",
			Name:        "TestStruct",
			URL:         "https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			IgnoreElements: []string{
				"script",
				"style",
				"link",
				"img",
				"footer",
				"header",
			},
			Fields: []Field{
				{
					Name:            "A",
					Type:            "string",
					Description:     "A description of the field",
					HeaderSelector:  "tr:nth-child(1) td:nth-child(1)",
					DataSelector:    "tr td:nth-child(1)",
					ControlSelector: "$text",
					MustBePresent:   "NCAA Codes",
				},
				{
					Name:            "B",
					Type:            "int",
					Description:     "A description of the field",
					HeaderSelector:  "tr:nth-child(1) td:nth-child(2)",
					DataSelector:    "tr td:nth-child(2)",
					ControlSelector: "$text",
					MustBePresent:   "NCAA Codes",
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestTestFilePromptArgs tests the TestFilePromptArgs struct.
func TestTestFilePromptArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		TestFilePromptArgs{
			Version:     "v0.0.0",
			Name:        "TestStruct",
			URL:         "https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			PackageName: "main",
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestSectionErrorPromptArgs tests the SectionErrorPromptArgs struct.
func TestSectionErrorPromptArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		sectionErrorArgs{
			Error: errors.New(
				"failed to get the content of the url: https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			),
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestIdentifyErrorArgs tests the IdentifyErrorArgs struct.
func TestIdentifyErrorArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		IdentifyErrorArgs{
			Error: errors.New(
				"failed to get the content of the url: https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			),
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestNewPromptIdentifyArgs tests the IdentifyPromptArgs struct with a single selector.
func TestNewPromptIdentifyArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		IdentifyArgs{
			URL:         "https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html",
			Content:     "<html><body><table><tr><td>a</td><td>b</td></tr><tr><td>1</td><td>2</td></tr></table></body></html>",
			NumSections: 3,
			Selectors: []master.Selector{
				{
					ID:         1,
					Value:      "html > body > table > tbody > tr > td:nth-child(1)",
					Occurances: 1,
					Context:    "<html></html>",
				},
			},
		},
	)
	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
}

// TestNewPromptPickSelectorArgs tests the NewPrompt function with a PickSelectorArgs struct.
func TestNewPromptPickSelectorArgs(t *testing.T) {
	a := assert.New(t)
	content, err := NewPrompt(
		PickSelectorArgs{
			Selectors: []master.Selector{
				{
					Value: "html > body > table#dataTable > tr:nth-child(1) > td:nth-child(1)",
				},
			},
			HTML: "<html><body><table id=\"dataTable\"><tr><td>a</td><td>b</td></tr><tr><td>1</td><td>2</td></tr></table></body></html>",
			Section: Section{
				Name:        "Test",
				Description: "Test Section",
				CSS:         "html > body > table#dataTable > tr:nth-child(1) > td:nth-child(1)",
			},
		},
	)

	a.NoError(err)
	a.NotEmpty(content)
	t.Logf("struct: %s", content)
	t.Fail()
}
