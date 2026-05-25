package app 

import (
	"context"
	t "dos/internal/common/types"
	"fmt"
)

func (app *App) ScaleObjects(ctx context.Context, objectID string, count int) error {
	err := app.MasterT.SetReplication(ctx, t.ObjectID(objectID), count)
	if err != nil {
		return err
	}

	fmt.Print(fmt.Sprintf("scaled object %s to %d replicas\n", objectID, count))
	return nil
}


