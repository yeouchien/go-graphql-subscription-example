import React, { Component } from 'react';
import { ApolloProvider } from 'react-apollo';
import { ApolloClient } from 'apollo-client'
import { InMemoryCache } from 'apollo-cache-inmemory';
import { HttpLink } from 'apollo-link-http'
import { WebSocketLink } from 'apollo-link-ws';
import { split } from 'apollo-link';
import { getMainDefinition } from 'apollo-utilities';
import Data from './Data'
import './App.css';

const httpLink = new HttpLink({
	uri: 'http://localhost:5000/graphql',
});

const wsLink = new WebSocketLink({
	uri: `ws://localhost:5000/subscriptions`,
	options: {
		reconnect: true
	}
});

const link = split(
	({ query }) => {
		const { kind, operation } = getMainDefinition(query);
		return kind === 'OperationDefinition' && operation === 'subscription';
	},
	wsLink,
	httpLink,
);


const client = new ApolloClient({
	link,
	cache: new InMemoryCache()
})

class App extends Component {
  render() {
    return (
		<ApolloProvider client={client}>
			<Data {...this.props} />
		</ApolloProvider>
    );
  }
}

export default App;
