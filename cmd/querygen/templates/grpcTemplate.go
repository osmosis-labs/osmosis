package templates

type GrpcTemplate struct {
	ProtoPath  string
	ClientPath string
	Queries    []GrpcQuery
}

type GrpcQuery struct {
	QueryFuncName   string
	QueryReqRepName string
	KeeperFuncName  string
}

func GrpcTemplateFromQueryYml(queryYml QueryYml) GrpcTemplate {
	GrpcQueries := []GrpcQuery{}
	for _, val := range queryYml.Queries {
		GrpcQueries = append(GrpcQueries, GrpcQuery{QueryFuncName: val.ProtoWrapper.QueryFuncName,
			QueryReqRepName: val.ProtoWrapper.QueryReqRepName,
			KeeperFuncName:  val.ProtoWrapper.KeeperFuncName})
	}

	return GrpcTemplate{
		ProtoPath:  queryYml.protoPath,
		ClientPath: queryYml.ClientPath,
		Queries:    GrpcQueries,
	}
}
