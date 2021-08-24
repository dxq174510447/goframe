package ctx

import (
	"context"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/util"
)

type StackCtx struct {
	context.Context
	stack *util.Stack
}

func (m *StackCtx) Value(key interface{}) interface{} {
	var key1 string
	if k, ok := key.(string); ok {
		key1 = k
	} else {
		key1 = fmt.Sprintf("%s", key)
	}

	v := m.stack.Get(key1)
	if v != nil {
		return v
	}
	return m.Context.Value(key)
}

func WithNewCtx(ctx context.Context) context.Context {
	mc := &StackCtx{
		Context: ctx,
		stack:   util.NewStack(),
	}
	return mc
}

func WithStackPush(ctx context.Context) context.Context {
	var mc *StackCtx
	if mapCtx, ok := ctx.(*StackCtx); ok {
		mc = mapCtx
	} else {
		mc = &StackCtx{
			Context: ctx,
			stack:   util.NewStack(),
		}
	}
	mc.stack.Push()
	return mc
}

func WithStackPop(ctx context.Context) context.Context {
	if mapCtx, ok := ctx.(*StackCtx); ok {
		mapCtx.stack.Pop()
	}
	return ctx
}

func WithStackPeek(ctx context.Context) map[string]interface{} {
	if mapCtx, ok := ctx.(*StackCtx); ok {
		return mapCtx.stack.Peek()
	}
	return nil
}

func WithValue(ctx context.Context, key string, val interface{}) context.Context {
	var mc *StackCtx
	if mapCtx, ok := ctx.(*StackCtx); ok {
		mc = mapCtx
	} else {
		mc = &StackCtx{
			Context: ctx,
			stack:   util.NewStack(),
		}
	}
	mc.stack.Set(key, val)
	return mc
}
