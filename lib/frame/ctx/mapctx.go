package ctx

import (
	"context"
	"fmt"
	"goframe/lib/frame/structure"
)

type StackCtx struct {
	context.Context
	stack *structure.Stack
}

func (m *StackCtx) Value(key interface{}) interface{} {
	var key1 string
	if k, ok := key.(string); ok {
		key1 = k
	} else {
		key1 = fmt.Sprintf("%s", key)
	}

	// 销毁
	if m.stack == nil {
		return nil
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
		stack:   structure.NewStack(),
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
			stack:   structure.NewStack(),
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

func Destory(ctx context.Context) {
	if mapCtx, ok := ctx.(*StackCtx); ok {
		mapCtx.stack.Destroy()
		mapCtx.stack = nil
	}
}

func WithValue(ctx context.Context, key string, val interface{}) context.Context {
	var mc *StackCtx
	if mapCtx, ok := ctx.(*StackCtx); ok {
		mc = mapCtx
	} else {
		mc = &StackCtx{
			Context: ctx,
			stack:   structure.NewStack(),
		}
	}
	mc.stack.Set(key, val)
	return mc
}
