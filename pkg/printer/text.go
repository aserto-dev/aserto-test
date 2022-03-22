package printer

import (
	"fmt"
	"io"
)

type Text struct {
	w io.Writer
}

func NewText(w io.Writer) *Text {
	return &Text{
		w: w,
	}
}

func (p *Text) Print(v []string) error {
	for _, s := range v {
		fmt.Fprintf(p.w, "%s\n", s)
	}
	fmt.Fprintln(p.w)
	return nil
}
