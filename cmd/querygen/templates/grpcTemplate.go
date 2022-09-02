package templates

type GrpcTemplate struct {
	ProtoPath  string
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
		ProtoPath:  queryYml.protoPath,
		ClientPath: queryYml.ClientPath,
		Queries:    GrpcQueries,
	}
}
