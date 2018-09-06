package hcl

import (
	"fmt"
)

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	// DiagInvalid is the invalid zero value of DiagnosticSeverity
	DiagInvalid DiagnosticSeverity = iota

	// DiagError indicates that the problem reported by a diagnostic prevents
	// further progress in parsing and/or evaluating the subject.
	DiagError

	// DiagWarning indicates that the problem reported by a diagnostic warrants
	// user attention but does not prevent further progress. It is most
	// commonly used for showing deprecation notices.
	DiagWarning
)

// Diagnostic represents information to be presented to a user about an
// error or anomoly in parsing or evaluating configuration.
type Diagnostic struct {
	Severity DiagnosticSeverity

	// Summary and Detail contain the English-language description of the
	// problem. Summary is a terse description of the general problem and
	// detail is a more elaborate, often-multi-sentence description of
	// the probem and what might be done to solve it.
	Summary string
	Detail  string

	// Subject and Context are both source ranges relating to the diagnostic.
	//
	// Subject is a tight range referring to exactly the construct that
	// is problematic, while Context is an optional broader range (which should
	// fully contain Subject) that ought to be shown around Subject when
	// generating isolated source-code snippets in diagnostic messages.
	// If Context is nil, the Subject is also the Context.
	//
	// Some diagnostics have no source ranges at all. If Context is set then
	// Subject should always also be set.
	Subject *Range
	Context *Range

	// For diagnostics that occur when evaluating an expression, Expression
	// may refer to that expression and EvalContext may point to the
	// EvalContext that was active when evaluating it. This may allow for the
	// inclusion of additional useful information when rendering a diagnostic
	// message to the user.
	//
	// It is not always possible to select a single EvalContext for a
	// diagnostic, and so in some cases this field may be nil even when an
	// expression causes a problem.
	//
	// EvalContexts form a tree, so the given EvalContext may refer to a parent
	// which in turn refers to another parent, etc. For a full picture of all
	// of the active variables and functions the caller must walk up this
	// chain, preferring definitions that are "closer" to the expression in
	// case of colliding names.
	Expression  Expression
	EvalContext *EvalContext
}

// Diagnostics is a list of Diagnostic instances.
type Diagnostics []*Diagnostic

// error implementation, so that diagnostics can be returned via APIs
// that normally deal in vanilla Go errors.
//
// This presents only minimal context about the error, for compatibility
// with usual expectations about how errors will present as strings.
func (d *Diagnostic) Error() string {
	if d.Subject == nil {
		return fmt.Sprintf("%s; %s", d.Summary, d.Detail)
	}
	return fmt.Sprintf("%s: %s; %s", d.Subject, d.Summary, d.Detail)
}

// error implementation, so that sets of diagnostics can be returned via
// APIs that normally deal in vanilla Go errors.
func (d Diagnostics) Error() string {
	count := len(d)
	switch {
	case count == 0:
		return "no diagnostics"
	case count == 1:
		return d[0].Error()
	default:
		return fmt.Sprintf("%s, and %d other diagnostic(s)", d[0].Error(), count-1)
	}
}

// Append appends a new error to a Diagnostics and return the whole Diagnostics.
//
// This is provided as a convenience for returning from a function that
// collects and then returns a set of diagnostics:
//
//     return nil, diags.Append(&hcl.Diagnostic{ ... })
//
// Note that this modifies the array underlying the diagnostics slice, so
// must be used carefully within a single codepath. It is incorrect (and rude)
// to extend a diagnostics created by a different subsystem.
func (d Diagnostics) Append(diag *Diagnostic) Diagnostics {
	return append(d, diag)
}

// Extend concatenates the given Diagnostics with the receiver and returns
// the whole new Diagnostics.
//
// This is similar to Append but accepts multiple diagnostics to add. It has
// all the same caveats and constraints.
func (d Diagnostics) Extend(diags Diagnostics) Diagnostics {
	return append(d, diags...)
}

// HasErrors returns true if the receiver contains any diagnostics of
// severity DiagError.
func (d Diagnostics) HasErrors() bool {
	for _, diag := range d {
		if diag.Severity == DiagError {
			return true
		}
	}
	return false
}

func (d Diagnostics) Errs() []error {
	var errs []error
	for _, diag := range d {
		if diag.Severity == DiagError {
			errs = append(errs, diag)
		}
	}

	return errs
}

// A DiagnosticWriter emits diagnostics somehow.
type DiagnosticWriter interface {
	WriteDiagnostic(*Diagnostic) error
	WriteDiagnostics(Diagnostics) error
}
