package router

type RouteImpl = routeImpl

func (r Router) FindRoutes(tokenInDenom, tokenOutDenom string, currentRoute Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]Route, error) {
	return r.findRoutes(tokenInDenom, tokenOutDenom, currentRoute, poolsUsed, previousTokenOutDenoms)
}
