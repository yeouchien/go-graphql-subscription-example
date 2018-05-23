import React, { Component } from 'react';
import { Query } from 'react-apollo';
import gql from 'graphql-tag';
import { Line } from 'react-chartjs-2';
import _ from 'lodash';
import moment from 'moment';


const DATA_SUBSCRIPTION = gql`
	subscription onNewData($input: NewDataInput!) {
		newData(input: $input) {
			timestamp
			value
			deviceId
		}
	}
`;

const DATA_QUERY = gql`
	query data {
		last {
			timestamp
			value
			deviceId
		}
	}
`;

const CHART_OPTIONS = {
  scales: {
    xAxes: [{
      type: 'time',
	    time: {
		    displayFormats: {
		      second: 'h:mm:ss',
		    },
	    },
      ticks: {
		    autoSkip: false,
        maxRotation: 90,
        minRotation: 90,
		    callback: function(value, index, values) {
		      return moment(value, 'h:mm:ss a').format('h:mm:ss a');
		    }
      },
    }],
    yAxes: [{
      ticks: {
        beginAtZero: true,
        min: 0,
        max: 100,
      },
    }],
  },
};

const CHART_DATA = {
  labels: ['Scatter'],
  datasets: [{
    label: 'Data',
    data: [],
    fill: false,
	  borderColor: 'rgb(255, 99, 132)',
	  backgroundColor: 'rgb(255, 99, 132)',
    borderWidth: 1,
    pointStyle: 'star',
    pointBorderWidth: 1,
  }]
};

class Chart extends Component {
	componentDidMount() {
		this.props.subscribeToNewData();
	}

	render() {
		if (!this.props.data || !this.props.data.last) {
			return null;
		}

		CHART_DATA.datasets[0].data = []

		_.forEach(this.props.data.last.slice(-10), (l) => {
			CHART_DATA.datasets[0].data.push({
				x: moment(l.timestamp).toDate(),
				y: l.value
			});
		});

		CHART_OPTIONS.scales.xAxes[0].time.min = CHART_DATA.datasets[0].data[0].x
		CHART_OPTIONS.scales.xAxes[0].time.max = CHART_DATA.datasets[0].data[CHART_DATA.datasets[0].data.length-1].x

		return (
			<Line data={CHART_DATA} options={CHART_OPTIONS} />
		)
	}
}

const DataChart = () => (
	<Query
	  query={DATA_QUERY}
	>
	  {({ subscribeToMore, ...result }) => (
	    <Chart
	      {...result}
	      subscribeToNewData={() => (
	        subscribeToMore({
	          document: DATA_SUBSCRIPTION,
			      variables: { input: { timestamp: moment().unix() } },
	          updateQuery: (prev, { subscriptionData }) => {
	            if (!subscriptionData.data) return prev;
	            const newData = subscriptionData.data.newData;
				      const newLast = [...prev.last, newData].slice(-10);

	            return Object.assign({}, prev, {
	                last: newLast
	            });
			      }
			    })
		    )}
	    />
	  )}
	</Query>
)

class Data extends Component {
	render() {
	  return (
	  	<div>
	  		<h2>My Data Chart</h2>
	  		<DataChart />
	  	</div>
	  );
	}
}

export default Data;
