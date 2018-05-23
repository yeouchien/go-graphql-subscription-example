package graph

import (
	tsdb "github.com/bluebreezecf/opentsdb-goclient/client"
	"github.com/bluebreezecf/opentsdb-goclient/config"
)

type Resolver struct {
	*newDataResolver
	opentsdbClient tsdb.Client
}

func NewResolver(opentsdbHost string) (*Resolver, error) {
	cl, err := tsdb.NewClient(config.OpenTSDBConfig{OpentsdbHost: opentsdbHost})
	if err != nil {
		return nil, err
	}

	return &Resolver{
		opentsdbClient: cl,
		newDataResolver: &newDataResolver{
			opentsdbClient: cl,
		},
	}, nil
}
