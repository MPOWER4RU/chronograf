import React, {PropTypes} from 'react';
import LayoutRenderer from '../components/LayoutRenderer';
import TimeRangeDropdown from '../../shared/components/TimeRangeDropdown';
import timeRanges from 'hson!../../shared/data/timeRanges.hson';
import {getMappings, getAppsForHosts, fetchLayouts} from '../apis';

export const HostPage = React.createClass({
  propTypes: {
    source: PropTypes.shape({
      links: PropTypes.shape({
        proxy: PropTypes.string.isRequired,
      }).isRequired,
    }),
    params: PropTypes.shape({
      hostID: PropTypes.string.isRequired,
    }).isRequired,
  },

  getInitialState() {
    const fifteenMinutesIndex = 1;

    return {
      layouts: [],
      timeRange: timeRanges[fifteenMinutesIndex],
    };
  },

  componentDidMount() {
    const hosts = {[this.props.params.hostID]: {name: this.props.params.hostID}};

    // fetching layouts and mappings can be done at the same time
    fetchLayouts().then(({data: {layouts}}) => {
      getMappings().then(({data: {mappings}}) => {
        getAppsForHosts(this.props.source.links.proxy, hosts, mappings).then((newHosts) => {
          const host = newHosts[this.props.params.hostID];
          const filteredLayouts = layouts.filter((layout) => {
            return host.apps && host.apps.includes(layout.app);
          });
          this.setState({layouts: filteredLayouts});
        });
      });
    });
  },

  handleChooseTimeRange({lower}) {
    const timeRange = timeRanges.find((range) => range.queryValue === lower);
    this.setState({timeRange});
  },

  renderLayout(layout) {
    const autoRefreshMs = 15000;
    const {timeRange} = this.state;
    const source = this.props.source.links.proxy;

    layout.cells.forEach((cell) => {
      cell.queries.forEach((q) => {
        q.text = q.query;
        q.database = q.db;
      });
    });

    return (
      <LayoutRenderer
        timeRange={timeRange.queryValue}
        cells={layout.cells}
        autoRefreshMs={autoRefreshMs}
        source={source}
        host={this.props.params.hostID}
      />
    );
  },

  render() {
    const hostID = this.props.params.hostID;
    const {layouts, timeRange} = this.state;

    return (
      <div className="host-dashboard hosts-page">
        <div className="enterprise-header hosts-dashboard-header">
          <div className="enterprise-header__container">
            <div className="enterprise-header__left">
              <h1>{hostID}</h1>
            </div>
            <div className="enterprise-header__right">
              <p>Uptime: <strong>2d 4h 33m</strong></p>
            </div>
            <div className="enterprise-header__right">
              <TimeRangeDropdown onChooseTimeRange={this.handleChooseTimeRange} selected={timeRange.inputValue} />
            </div>
          </div>
        </div>
        <div className="container-fluid hosts-dashboard">
          {
            layouts.map((layout) => {
              return (
                <div key={layout.app} className="row">
                  {this.renderLayout(layout)}
                </div>
              );
            })
          }
        </div>
      </div>
    );
  },
});

export default HostPage;
