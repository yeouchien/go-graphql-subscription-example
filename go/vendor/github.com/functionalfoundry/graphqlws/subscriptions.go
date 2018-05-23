package graphqlws

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// SubscriptionSendDataFunc is a function that sends updated data
// for a specific subscription to the corresponding subscriber.
type SubscriptionSendDataFunc func(*DataMessagePayload)

// Subscription holds all information about a GraphQL subscription
// made by a client, including a function to send data back to the
// client when there are updates to the subscription query result.
type Subscription struct {
	stopCh        chan struct{}
	ID            string
	Query         string
	Variables     map[string]interface{}
	OperationName string
	Fields        []string
	Connection    Connection
	SendData      SubscriptionSendDataFunc
}

func (s *Subscription) StopCh() <-chan struct{} {
	return s.stopCh
}

// ConnectionSubscriptions defines a map of all subscriptions of
// a connection by their IDs.
type ConnectionSubscriptions map[string]*Subscription

// Subscriptions defines a map of connections to a map of
// subscription IDs to subscriptions.
type Subscriptions map[Connection]ConnectionSubscriptions

// SubscriptionManager provides a high-level interface to managing
// and accessing the subscriptions made by GraphQL WS clients.
type SubscriptionManager interface {
	// Subscriptions returns all registered subscriptions, grouped
	// by connection.
	Subscriptions() Subscriptions

	// AddSubscription adds a new subscription to the manager.
	AddSubscription(Connection, *Subscription) []error

	// RemoveSubscription removes a subscription from the manager.
	RemoveSubscription(Connection, *Subscription)

	// RemoveSubscriptions removes all subscriptions of a client connection.
	RemoveSubscriptions(Connection)
}

/**
 * The default implementation of the SubscriptionManager interface.
 */

type subscriptionListener func(*Subscription)

type subscriptionManager struct {
	subscriptionListeners []subscriptionListener
	subscriptions         Subscriptions
	logger                *log.Entry
}

// NewSubscriptionManager creates a new subscription manager.
func NewSubscriptionManager(subscriptionListeners ...subscriptionListener) SubscriptionManager {
	manager := new(subscriptionManager)
	manager.subscriptionListeners = subscriptionListeners
	manager.subscriptions = make(Subscriptions)
	manager.logger = NewLogger("subscriptions")
	return manager
}

func (m *subscriptionManager) Subscriptions() Subscriptions {
	return m.subscriptions
}

func (m *subscriptionManager) AddSubscription(
	conn Connection,
	subscription *Subscription,
) []error {
	m.logger.WithFields(log.Fields{
		"conn":         conn.ID(),
		"subscription": subscription.ID,
	}).Info("Add subscription")

	if errors := validateSubscription(subscription); len(errors) > 0 {
		m.logger.WithField("errors", errors).Warn("Failed to add invalid subscription")
		return errors
	}

	// Allocate the connection's map of subscription IDs to
	// subscriptions on demand
	if m.subscriptions[conn] == nil {
		m.subscriptions[conn] = make(ConnectionSubscriptions)
	}

	// Add the subscription if it hasn't already been added
	if m.subscriptions[conn][subscription.ID] != nil {
		m.logger.WithFields(log.Fields{
			"conn":         conn.ID(),
			"subscription": subscription.ID,
		}).Warn("Cannot register subscription twice")
		return []error{errors.New("Cannot register subscription twice")}
	}

	m.subscriptions[conn][subscription.ID] = subscription
	for _, f := range m.subscriptionListeners {
		go f(subscription)
	}

	return nil
}

func (m *subscriptionManager) RemoveSubscription(
	conn Connection,
	subscription *Subscription,
) {
	m.logger.WithFields(log.Fields{
		"conn":         conn.ID(),
		"subscription": subscription.ID,
	}).Info("Remove subscription")

	// close channel
	if _, okConn := m.subscriptions[conn]; okConn {
		if s, okSub := m.subscriptions[conn][subscription.ID]; okSub {
			close(s.stopCh)
		}
	}

	// Remove the subscription from its connections' subscription map
	delete(m.subscriptions[conn], subscription.ID)

	// Remove the connection as well if there are no subscriptions left
	if len(m.subscriptions[conn]) == 0 {
		delete(m.subscriptions, conn)
	}
}

func (m *subscriptionManager) RemoveSubscriptions(conn Connection) {
	m.logger.WithFields(log.Fields{
		"conn": conn.ID(),
	}).Info("Remove subscriptions")

	// Only remove subscriptions if we know the connection
	if m.subscriptions[conn] != nil {
		// Remove subscriptions one by one
		for opID := range m.subscriptions[conn] {
			m.RemoveSubscription(conn, m.subscriptions[conn][opID])
		}

		// Remove the connection's subscription map altogether
		delete(m.subscriptions, conn)
	}
}

func validateSubscription(s *Subscription) []error {
	errs := []error{}

	if s.ID == "" {
		errs = append(errs, errors.New("Subscription ID is empty"))
	}

	if s.Connection == nil {
		errs = append(errs, errors.New("Subscription is not associated with a connection"))
	}

	if s.Query == "" {
		errs = append(errs, errors.New("Subscription query is empty"))
	}

	if s.SendData == nil {
		errs = append(errs, errors.New("Subscription has no SendData function set"))
	}

	return errs
}
