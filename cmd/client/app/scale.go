package app 

import (
	"context"

	t "dos/internal/common/types"
)

func (app *App) ScaleObject(ctx context.Context, objectID string, count int) error {
	err := app.MasterT().SetReplication(ctx, t.ObjectID(objectID), count)
	if err != nil {
		return err
	}
	return nil
}


