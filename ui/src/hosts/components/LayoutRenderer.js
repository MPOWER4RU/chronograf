import React, {PropTypes} from 'react';
import AutoRefresh from 'shared/components/AutoRefresh';
import LineGraph from 'shared/components/LineGraph';
import ReactGridLayout from 'react-grid-layout';
import _ from 'lodash';

const RefreshingLineGraph = AutoRefresh(LineGraph);

export const LayoutRenderer = React.createClass({
  propTypes: {
    cells: PropTypes.arrayOf(
      PropTypes.shape({
        queries: PropTypes.arrayOf(
          PropTypes.shape({
            rp: PropTypes.string.isRequired,
            text: PropTypes.string.isRequired,
            database: PropTypes.string.isRequired,
          }).isRequired
        ).isRequired,
        x: PropTypes.number.isRequired,
        y: PropTypes.number.isRequired,
        w: PropTypes.number.isRequired,
        h: PropTypes.number.isRequired,
        i: PropTypes.string.isRequired,
        name: PropTypes.string.isRequired,
      }).isRequired
    ),
    autoRefreshMs: PropTypes.number.isRequired,
    host: PropTypes.string.isRequired,
    source: PropTypes.string,
  },

  getInitialState() {
    return ({
      layout: _.without(this.props.cells, ['queries']),
    });
  },

  generateGraphs() {
    return this.props.cells.map((cell) => {
      const qs = cell.queries.map((q) => {
        _.merge(q, {host: this.props.source});
        q.text += ` where host = '${this.props.host}' and time > now() - 15m`;
        return q;
      });
      return (
        <div key={cell.i}>
          <h2 className="hosts-graph-heading">{cell.name}</h2>
          <div className="hosts-graph graph-panel__graph-container">
            <RefreshingLineGraph
              queries={qs}
              autoRefresh={this.props.autoRefreshMs}
            />
          </div>
        </div>
      );
    });
  },

  render() {
    return (
      <ReactGridLayout layout={this.state.layout} isDraggable={false} isResizable={false} cols={12} rowHeight={30} width={1200}>
        {this.generateGraphs()}
      </ReactGridLayout>
    );
  },
});

export default LayoutRenderer;
