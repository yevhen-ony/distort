package render

import (
	"dos/cmd/client/app"
	"errors"
	"fmt"
	"io"
	"strings"
)

type TextRender struct {
	out io.Writer
}

func NewTextRender(out io.Writer) (*TextRender, error) {
	if out == nil {
		return nil, errors.New("missing out")
	}
	render := &TextRender{
		out: out,
	}
	return render, nil
}

func (r *TextRender) Error(opName string, opErr error) error {
	_, err := fmt.Fprintf(r.out,
		"%s:\n\t * error: %s\n",
		strings.ToUpper(opName),
		opErr.Error(),
	)
	return err
}

func (r *TextRender) Ping(ping *app.PingResult) error {
	b := &strings.Builder{}

	fmt.Fprintln(b, "PING:")
	fmt.Fprintf(b, "\t * address  : %s\n", ping.Address)
	fmt.Fprintf(b, "\t * status   : %s\n", ping.Status)
	fmt.Fprintf(b, "\t * component: %s\n", ping.Component)

	_, err := fmt.Fprint(r.out, b.String())
	return err
}


