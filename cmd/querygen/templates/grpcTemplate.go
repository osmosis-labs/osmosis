package templates

import "sort"

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
	sort.Slice(GrpcQueries, func(i, j int) bool {
		return GrpcQueries[i].QueryName > GrpcQueries[j].QueryName
	})
	return GrpcTemplate{
		ProtoPath:  queryYml.protoPath,
		ClientPath: queryYml.ClientPath,
		Queries:    GrpcQueries,
	}
}
