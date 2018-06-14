package graph

import (
	"context"
	"log"
	"strconv"
	"time"

	tsdb "github.com/bluebreezecf/opentsdb-goclient/client"
	graphql "github.com/graph-gophers/graphql-go"
)

type data struct {
	timestamp graphql.Time
	value     float64
	deviceID  string
}

type dataResolver struct {
	d *data
}

func (d *dataResolver) Timestamp() graphql.Time {
	return d.d.timestamp
}

func (d *dataResolver) Value() float64 {
	return d.d.value
}

func (d *dataResolver) DeviceID() string {
	return d.d.deviceID
}

type createDataInput struct {
	Timestamp graphql.Time `json:"timestamp"`
	Value     float64      `json:"value"`
	DeviceID  string       `json:"device_id"`
}

func (r *Resolver) CreateData(ctx context.Context, args struct {
	Input createDataInput
}) (*dataResolver, error) {
	timestamp := args.Input.Timestamp.Unix()
	value := args.Input.Value
	deviceID := args.Input.DeviceID

	dps := []tsdb.DataPoint{
		tsdb.DataPoint{
			Metric:    "graphql",
			Timestamp: timestamp,
			Value:     value,
			Tags: map[string]string{
				"device_id": deviceID,
			},
		},
	}

	_, err := r.opentsdbClient.Put(dps, tsdb.PutRespWithSummary)
	if err != nil {
		return nil, err
	}

	d := &data{
		timestamp: args.Input.Timestamp,
		value:     args.Input.Value,
		deviceID:  args.Input.DeviceID,
	}

	return &dataResolver{d: d}, nil
}

func (r *Resolver) Last(ctx context.Context) ([]*dataResolver, error) {
	res, err := r.opentsdbClient.Query(tsdb.QueryParam{
		Start: time.Now().Add(-5 * time.Second).Unix(),
		Queries: []tsdb.SubQuery{
			tsdb.SubQuery{
				Aggregator: "avg",
				Metric:     "graphql",
				Tags: map[string]string{
					"device_id": "device-id",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var dcs []*dataResolver
	for _, qr := range res.QueryRespCnts {
		dps := qr.GetDataPoints()
		for _, dp := range dps {
			d := &data{
				timestamp: graphql.Time{time.Unix(dp.Timestamp, 0)},
				value:     dp.Value.(float64),
				deviceID:  dp.Tags["device_id"],
			}
			dcs = append(dcs, &dataResolver{d})
		}
	}

	if len(dcs) == 0 {
		d := &data{
			timestamp: graphql.Time{time.Now()},
			value:     0,
			deviceID:  "device-id",
		}
		dcs = append(dcs, &dataResolver{d})
	}

	return dcs, nil
}

type newDataResolver struct {
	opentsdbClient tsdb.Client
	newDataEvents  chan *newDataEvent
}

type newDataEvent struct {
	timestamp graphql.Time
	value     float64
	deviceID  string
	err       error
}

type newDataInput struct {
	Timestamp graphql.Time `json:"timestamp"`
}

func (r *newDataResolver) NewData(ctx context.Context, args struct {
	Input newDataInput
}) <-chan *newDataEvent {
	c := make(chan *newDataEvent)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				res, err := r.opentsdbClient.QueryLast(tsdb.QueryLastParam{
					BackScan: 24, // 24 hours
					Queries: []tsdb.SubQueryLast{
						tsdb.SubQueryLast{
							Metric: "graphql",
							Tags: map[string]string{
								"device_id": "device-id",
							},
						},
					},
				})
				if err != nil {
					log.Printf("error querying last: %v", err)
					return
				}

				for _, dp := range res.QueryRespCnts {
					value, err := strconv.ParseFloat(dp.Value, 64)
					if err != nil {
						log.Printf("error parsing float: %v", err)
						return
					}

					c <- &newDataEvent{
						timestamp: graphql.Time{time.Unix(dp.Timestamp/1000, 0)},
						value:     value,
						deviceID:  dp.Tags["device_id"],
					}
				}
			}
		}
	}()

	return c
}

func (r *newDataEvent) Timestamp() graphql.Time {
	return r.timestamp
}

func (r *newDataEvent) Value() float64 {
	return r.value
}

func (r *newDataEvent) DeviceID() string {
	return r.deviceID
}
