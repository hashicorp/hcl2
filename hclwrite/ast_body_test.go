package hclwrite

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestBodyGetAttribute(t *testing.T) {
	tests := []struct {
		src  string
		name string
		want Tokens
	}{
		{
			"",
			"a",
			nil,
		},
		{
			"a = 1\n",
			"a",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'1'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"a",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'1'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = 2\nc = 3\n",
			"c",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'c'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'3'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\n# b is a b\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					// Recognized as a lead comment and so attached to the attribute
					Type:         hclsyntax.TokenComment,
					Bytes:        []byte("# b is a b\n"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\n# not attached to a or b\n\nb = 2\nc = 3\n",
			"b",
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte{'2'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s in %s", test.name, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			attr := f.Body().GetAttribute(test.name)
			if attr == nil {
				if test.want != nil {
					t.Fatal("attribute not found, but want it to exist")
				}
			} else {
				if test.want == nil {
					t.Fatal("attribute found, but expecting not found")
				}

				got := attr.BuildTokens(nil)
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
				}
			}
		})
	}
}

func TestBodyGetBlock(t *testing.T) {
	src := `a = "b"
service {
  attr0 = "val0"
}
service "label1" {
  attr1 = "val1"
}
service "label1" "label2" {
  attr2 = "val2"
}
parent {
  attr3 = "val3"
  child {
    attr4 = "val4"
  }
}
`

	tests := []struct {
		src      string
		typeName string
		labels   []string
		want     string
	}{
		{
			src,
			"service",
			[]string{},
			`service {
  attr0 = "val0"
}
`,
		},
		{
			src,
			"service",
			[]string{"label1"},
			`service "label1" {
  attr1 = "val1"
}
`,
		},
		{
			src,
			"service",
			[]string{"label1", "label2"},
			`service "label1" "label2" {
  attr2 = "val2"
}
`,
		},
		{
			src,
			"parent",
			[]string{},
			`parent {
  attr3 = "val3"
  child {
    attr4 = "val4"
  }
}
`,
		},
		{
			src,
			"hoge",
			[]string{},
			"",
		},
		{
			src,
			"hoge",
			[]string{"label1"},
			"",
		},
		{
			src,
			"service",
			[]string{"label2"},
			"",
		},
		{
			src,
			"service",
			[]string{"label2", "label1"},
			"",
		},
		{
			src,
			"child",
			[]string{},
			"",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.typeName, strings.Join(test.labels, " ")), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			block := f.Body().GetBlock(test.typeName, test.labels)
			if block == nil {
				if test.want != "" {
					t.Fatal("block not found, but want it to exist")
				}
			} else {
				if test.want == "" {
					t.Fatal("block found, but expecting not found")
				}

				got := string(block.BuildTokens(nil).Bytes())
				if got != test.want {
					t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.want)
				}
			}
		})
	}
}

func TestBodySetAttributeValue(t *testing.T) {
	tests := []struct {
		src  string
		name string
		val  cty.Value
		want Tokens
	}{
		{
			"",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"b = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("false"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"a = 1\nb = false\n",
			"a",
			cty.True,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("true"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'b'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("false"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s = %#v in %s", test.name, test.val, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			f.Body().SetAttributeValue(test.name, test.val)
			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestBodySetAttributeTraversal(t *testing.T) {
	tests := []struct {
		src  string
		name string
		trav string
		want Tokens
	}{
		{
			"",
			"a",
			`b`,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("b"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"",
			"a",
			`b.c.d`,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("b"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenDot,
					Bytes:        []byte("."),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("c"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenDot,
					Bytes:        []byte("."),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("d"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"",
			"a",
			`b[0]`,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("b"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOBrack,
					Bytes:        []byte("["),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte("0"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrack,
					Bytes:        []byte("]"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"",
			"a",
			`b[0].c`,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte{'a'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte{'='},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("b"),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOBrack,
					Bytes:        []byte("["),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNumberLit,
					Bytes:        []byte("0"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrack,
					Bytes:        []byte("]"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenDot,
					Bytes:        []byte("."),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte("c"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s = %s in %s", test.name, test.trav, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			traversal, diags := hclsyntax.ParseTraversalAbs([]byte(test.trav), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics from traversal")
			}

			f.Body().SetAttributeTraversal(test.name, traversal)
			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}

func TestBodySetAttributeValueInBlock(t *testing.T) {
	src := `service "label1" {
  attr1 = "val1"
}
`
	tests := []struct {
		src      string
		typeName string
		labels   []string
		attr     string
		val      cty.Value
		want     string
	}{
		{
			src,
			"service",
			[]string{"label1"},
			"attr1",
			cty.StringVal("updated1"),
			`service "label1" {
  attr1 = "updated1"
}
`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s = %#v in %s %s", test.attr, test.val, test.typeName, strings.Join(test.labels, " ")), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			b := f.Body().GetBlock(test.typeName, test.labels)
			b.Body().SetAttributeValue(test.attr, test.val)
			tokens := f.BuildTokens(nil)
			format(tokens)
			got := string(tokens.Bytes())
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s\n", got, test.want)
			}
		})
	}
}

func TestBodySetAttributeValueInNestedBlock(t *testing.T) {
	src := `parent {
  attr1 = "val1"
  child {
    attr2 = "val2"
  }
}
`
	tests := []struct {
		src            string
		parentTypeName string
		childTypeName  string
		attr           string
		val            cty.Value
		want           string
	}{
		{
			src,
			"parent",
			"child",
			"attr2",
			cty.StringVal("updated2"),
			`parent {
  attr1 = "val1"
  child {
    attr2 = "updated2"
  }
}
`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s = %#v in %s in %s", test.attr, test.val, test.childTypeName, test.parentTypeName), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			parent := f.Body().GetBlock(test.parentTypeName, []string{})
			child := parent.Body().GetBlock(test.childTypeName, []string{})
			child.Body().SetAttributeValue(test.attr, test.val)
			tokens := f.BuildTokens(nil)
			format(tokens)
			got := string(tokens.Bytes())
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s\n", got, test.want)
			}
		})
	}
}

func TestBodyAppendBlock(t *testing.T) {
	tests := []struct {
		src       string
		blockType string
		labels    []string
		want      Tokens
	}{
		{
			"",
			"foo",
			nil,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"",
			"foo",
			[]string{"bar"},
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(`bar`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"",
			"foo",
			[]string{"bar", "baz"},
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(`bar`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(`baz`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
		{
			"bar {}\n",
			"foo",
			nil,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`bar`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte{'{'},
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte{'}'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %#v in %s", test.blockType, test.blockType, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			f.Body().AppendNewBlock(test.blockType, test.labels)
			got := f.BuildTokens(nil)
			format(got)
			if !reflect.DeepEqual(got, test.want) {
				diff := cmp.Diff(test.want, got)
				t.Errorf("wrong result\ngot:  %s\nwant: %s\ndiff:\n%s", spew.Sdump(got), spew.Sdump(test.want), diff)
			}
		})
	}
}
