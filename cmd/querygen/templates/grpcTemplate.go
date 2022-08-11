package templates

type GrpcTemplate struct {
	ClientPath string
	Queries    []GrpcQuery
}

type GrpcQuery struct {
	QueryName string
}

func GrpcTemplateFromQueryYml(queryYml QueryYml) GrpcTemplate {
	GrpcQueries := []GrpcQuery{}
	for queryName := range queryYml.Queries {
		GrpcQueries = append(GrpcQueries, GrpcQuery{QueryName: queryName})
	}
	return GrpcTemplate{
		ClientPath: queryYml.ClientPath,
		Queries:    GrpcQueries,
	}
}
