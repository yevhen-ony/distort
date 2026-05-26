package object

import "context"

type ObjectCommandApplier interface {
  	ApplyObjectCommand(context.Context, ObjectCommand) error
}
