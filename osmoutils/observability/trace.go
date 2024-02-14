package observability

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/osmosis-labs/osmosis/x/epochs/types"
)

// InitSDKCtxWithSpan initializes sdk.Context with span of the given name created from the given tracer.
// The caller must call span.End() on the span.
func InitSDKCtxWithSpan(goCtx context.Context, tracer trace.Tracer, spanName string) (sdk.Context, trace.Span) {
	goCtx, span := tracer.Start(goCtx, spanName)
	goCtx = trace.ContextWithSpan(goCtx, span)
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx = ctx.WithContext(goCtx)

	return ctx, span
}

// GetSpanFromSDKContext
func GetSpanFromSDKContext(ctx sdk.Context) trace.Span {
	goCtx := sdk.WrapSDKContext(ctx)
	span := trace.SpanFromContext(goCtx)

	return span
}

// EmitSwapEvent emits a swap event for tracing
func EmitSwapEvent(span trace.Span, poolID uint64, tokenIn, tokenOut sdk.Coin) {
	span.AddEvent("swap",
		trace.WithAttributes(
			attribute.Int64("pool_id", int64(poolID)),
			attribute.String("pool_type", types.ModuleName),
			attribute.Stringer("token_in", tokenIn),
			attribute.Stringer("token_out", tokenOut),
		),
	)
}
