// Adopted from https://github.com/open-policy-agent/opa/blob/main/tester/reporter.go
//
// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.
//
package tester

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Reporter defines the interface for reporting test results.
type Reporter interface {
	Report(ch []*TestResult) error
}

// PrettyReporter reports test results in a simple human readable format.
type PrettyReporter struct {
	Output      io.Writer
	Verbose     bool
	FailureLine bool
}

// Report prints the test report to the reporter's output.
func (r PrettyReporter) Report(ch []*TestResult) error {
	dirty := false
	var pass, fail, skip, errs int
	results := make([]*TestResult, len(ch))
	var failures []*TestResult

	for i, tr := range ch {
		if tr.Pass() {
			pass++
		} else if tr.Skip {
			skip++
		} else if tr.Error != nil {
			errs++
		} else if tr.Fail {
			fail++
			failures = append(failures, tr)
		}
		results[i] = tr
	}

	if fail > 0 && r.Verbose {
		fmt.Fprintln(r.Output, "FAILURES")
		r.hl()

		for _, failure := range failures {
			fmt.Fprintln(r.Output, failure)
			fmt.Fprintln(r.Output)
			for _, l := range failure.Output {
				fmt.Fprintln(newIndentingWriter(r.Output), strings.TrimSpace(l))
			}
			fmt.Fprintln(r.Output)
		}

		fmt.Fprintln(r.Output, "SUMMARY")
		r.hl()
	}

	// Report individual tests.
	for _, tr := range results {
		dirty = true
		fmt.Fprintln(r.Output, tr)

		if tr.Error != nil {
			fmt.Fprintf(r.Output, "  %v\n", tr.Error)
		}
	}

	// Report summary of test.
	if dirty {
		r.hl()
	}

	total := pass + fail + skip + errs

	if pass != 0 {
		fmt.Fprintln(r.Output, "PASS:", fmt.Sprintf("%d/%d", pass, total))
	}

	if fail != 0 {
		fmt.Fprintln(r.Output, "FAIL:", fmt.Sprintf("%d/%d", fail, total))
	}

	if skip != 0 {
		fmt.Fprintln(r.Output, "SKIPPED:", fmt.Sprintf("%d/%d", skip, total))
	}

	if errs != 0 {
		fmt.Fprintln(r.Output, "ERROR:", fmt.Sprintf("%d/%d", errs, total))
	}

	return nil
}

func (r PrettyReporter) hl() {
	fmt.Fprintln(r.Output, strings.Repeat("-", 80))
}

// JSONReporter reports test results as array of JSON objects.
type JSONReporter struct {
	Output io.Writer
}

// Report prints the test report to the reporter's output.
func (r JSONReporter) Report(ch []*TestResult) error {
	var report []*TestResult
	report = append(report, ch...)

	bs, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(r.Output, string(bs))
	return nil
}

type indentingWriter struct {
	w io.Writer
}

func newIndentingWriter(w io.Writer) indentingWriter {
	return indentingWriter{
		w: w,
	}
}

func (w indentingWriter) Write(bs []byte) (int, error) {
	var written int
	// insert indentation at the start of every line.
	indent := true
	for _, b := range bs {
		if indent {
			wrote, err := w.w.Write([]byte("  "))
			if err != nil {
				return written, err
			}
			written += wrote
		}
		wrote, err := w.w.Write([]byte{b})
		if err != nil {
			return written, err
		}
		written += wrote
		indent = b == '\n'
	}
	return written, nil
}
