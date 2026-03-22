package format

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type Table struct {
	w       *tabwriter.Writer
	headers []string
}

func NewTable(headers ...string) *Table {
	return NewTableWriter(os.Stdout, headers...)
}

func NewTableWriter(out io.Writer, headers ...string) *Table {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	t := &Table{w: w, headers: headers}
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	dashes := make([]string, len(headers))
	for i, h := range headers {
		dashes[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(w, strings.Join(dashes, "\t"))
	return t
}

func (t *Table) Row(values ...string) {
	fmt.Fprintln(t.w, strings.Join(values, "\t"))
}

func (t *Table) Flush() {
	t.w.Flush()
}
