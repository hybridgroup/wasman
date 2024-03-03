package segments

import (
	"fmt"

	"github.com/hybridgroup/wasman/expr"
	"github.com/hybridgroup/wasman/types"
	"github.com/hybridgroup/wasman/utils"
)

// GlobalSegment is one unit of the wasm.Module's GlobalSection
type GlobalSegment struct {
	Type *types.GlobalType
	Init *expr.Expression
}

// ReadGlobalSegment reads one GlobalSegment from the io.Reader
func ReadGlobalSegment(r utils.Reader) (*GlobalSegment, error) {
	gt, err := types.ReadGlobalType(r)
	if err != nil {
		return nil, fmt.Errorf("read global type: %w", err)
	}

	init, err := expr.ReadExpression(r)
	if err != nil {
		return nil, fmt.Errorf("get init expression: %w", err)
	}

	return &GlobalSegment{
		Type: gt,
		Init: init,
	}, nil
}
