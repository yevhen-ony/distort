package app

import (
	"context"
	"fmt"
	"strings"
)

func (app *App) Ping(ctx context.Context, addr string) {
	res, err := app.HealthT.Ready(ctx, addr)
	
	if err != nil {
		fmt.Println(makePingOutput(addr, "not ready", "unknown"))
	}
	fmt.Println(makePingOutput(addr, "ready", res.Component))
}

func makePingOutput(addr, status, component string) string {
	b := &strings.Builder{}

	fmt.Fprintln(b, "PING:")
	fmt.Fprintf(b, "\t * address  : %s\n", addr)
	fmt.Fprintf(b, "\t * status   : %s\n", status)
	fmt.Fprintf(b, "\t * component: %s\n", component)  
	return b.String()
}


