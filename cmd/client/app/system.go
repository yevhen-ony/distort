package app

import (
	"context"
	"fmt"
	"strings"
)

func (app *App) DiscoverActiveMaster(ctx context.Context) error {
	ref, err := app.MasterT().DiscoverMaster(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, "%-20s %-20s\n", "MASTER_ID", "MASTER_ADDR")
	fmt.Fprintf(b, "%-20s %-20s\n", ref.ID, ref.Addr)

	fmt.Print(b.String())
	return nil
}

