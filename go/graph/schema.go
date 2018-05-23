package graph

var Schema = `scalar Time

schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

type Query {
	last: [Data!]!
}

type Mutation {
	createData(
		input: CreateDataInput!
	): Data!
}

type Subscription {
	newData(
		input: NewDataInput!
	): NewDataEvent!
}

type NewDataEvent {
	timestamp: Time!
    value: Float!
    deviceId: String!
}

type Data {
	timestamp: Time!
    value: Float!
    deviceId: String!
}

type DataCollection {
	nodes: [Data!]!
}

input CreateDataInput {
	timestamp: Time!
    value: Float!
    deviceId: String!
}

input NewDataInput {
	timestamp: Time!
}`
