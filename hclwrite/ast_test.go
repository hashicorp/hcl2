package hclwrite

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestBodyFindAttribute(t *testing.T) {
	tests := []struct {
		src  string
		name string
		want *TokenSeq
	}{
		{
			"",
			"a",
			nil,
		},
		{
			"a = 1\n",
			"a",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'a'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"a",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'a'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"b",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'b'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"c",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'c'},
					},
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

			attr := f.Body.FindAttribute(test.name)
			if attr == nil {
				if test.want != nil {
					t.Errorf("attribute found, but expecting not found")
				}
			} else {
				got := attr.NameTokens
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
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
		want string
	}{
		{
			"",
			"a",
			cty.True,
			"a = true\n",
		},
		{
			"b = 0\n",
			"a",
			cty.True,
			"b = 0\na = true\n",
		},
		{
			"a = 0\n",
			"a",
			cty.True,
			"a = true\n",
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

			f.Body.SetAttributeValue(test.name, test.val)
			var gotBuf bytes.Buffer
			f.WriteTo(&gotBuf)
			got := gotBuf.String()
			want := test.want

			if !cmp.Equal(got, want) {
				t.Logf(spew.Sdump(f))
				t.Errorf("wrong result\n%s", cmp.Diff(want, got))
			}
		})
	}
}
