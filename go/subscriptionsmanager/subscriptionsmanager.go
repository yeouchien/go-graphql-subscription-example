package subscriptionsmanager

import (
	"context"
	"log"

	"github.com/functionalfoundry/graphqlws"
	graphql "github.com/graph-gophers/graphql-go"
)

type SubscriptionsManager struct {
	graphqlws.SubscriptionManager
	ctx    context.Context
	subCh  chan *graphqlws.Subscription
	schema *graphql.Schema
}

func New(ctx context.Context, c chan *graphqlws.Subscription, s *graphql.Schema) *SubscriptionsManager {
	return &SubscriptionsManager{
		SubscriptionManager: graphqlws.NewSubscriptionManager(
			func(subscription *graphqlws.Subscription) {
				log.Printf("new subscription")
				c <- subscription
			},
		),
		ctx:    ctx,
		subCh:  c,
		schema: s,
	}
}

func (sm *SubscriptionsManager) Start() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case sub := <-sm.subCh:
			go func() {
				ctx, cancel := context.WithCancel(sm.ctx)
				c, err := sm.schema.Subscribe(ctx, sub.Query, sub.OperationName, sub.Variables)
				if err != nil {
					log.Printf("error subscribing: %v", err)
					return
				}

				for {
					select {
					case <-sm.ctx.Done():
						cancel()
						return
					case <-sub.StopCh():
						log.Printf("shutdown upstream sub %v", sub.ID)
						cancel()
						return
					case resp := <-c:
						if resp != nil {
							errs := []error{}
							for _, e := range resp.Errors {
								errs = append(errs, e)
							}

							data := graphqlws.DataMessagePayload{
								Data:   resp.Data,
								Errors: errs,
							}

							sub.SendData(&data)
						}
					}
				}
			}()

		}

	}
}
