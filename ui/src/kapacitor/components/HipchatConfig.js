import React, {PropTypes} from 'react';

const HipchatConfig = React.createClass({
  propTypes: {
    config: PropTypes.shape({
      options: PropTypes.shape({
        global: PropTypes.bool.isRequired,
        room: PropTypes.string.isRequired,
        'state-changes-only': PropTypes.bool.isRequired,
        token: PropTypes.bool.isRequired,
        url: PropTypes.string.isRequired,
      }).isRequired,
    }).isRequired,
    onSave: PropTypes.func.isRequired,
  },

  handleSaveAlert(e) {
    e.preventDefault();

    const properties = {
      room: this.room.value,
      url: this.url.value,
      token: this.token.value,
      'state-changes-only': this.stateChangesOnly.checked,
      global: this.global.checked,
    };

    this.props.onSave(properties);
  },

  render() {
    const {options} = this.props.config;
    const stateChangesOnly = options['state-changes-only'];
    const {url, global, room, token} = options;

    return (
      <div className="panel-body">
        <h4 className="text-center">VictorOps Alert</h4>
        <br/>
        <form onSubmit={this.handleSaveAlert}>
          <div className="row">
            <div className="col-xs-7 col-sm-8 col-sm-offset-2">
              <p>
                Have alerts sent to HipChat
              </p>

              <div className="form-group">
                <label htmlFor="url">HipChat URL</label>
                <input className="form-control" id="url" type="text" ref={(r) => this.url = r} defaultValue={url || ''}></input>
              </div>

              <div className="form-group">
                <label htmlFor="room">Room</label>
                <input className="form-control" id="room" type="text" ref={(r) => this.room = r} defaultValue={room || ''}></input>
              </div>

              <div className="form-group">
                <label htmlFor="token">Token</label>
                <input className="form-control" id="token" type="text" ref={(r) => this.token = r} defaultValue={token || ''}></input>
                <span>Note: a value of <code>true</code> indicates the HipChat token has been set</span>
              </div>

              <div className="form-group col-xs-12">
                <div className="checkbox">
                  <label>
                    <input type="checkbox" defaultChecked={global} ref={(r) => this.global = r} />
                    Send all alerts without marking them explicitly in TICKscript
                  </label>
                </div>
              </div>

              <div className="form-group col-xs-12">
                <div className="checkbox">
                  <label>
                    <input type="checkbox" defaultChecked={stateChangesOnly} ref={(r) => this.stateChangesOnly = r} />
                    Send alerts on state change only
                  </label>
                </div>
              </div>
            </div>
          </div>

          <hr />
          <div className="row">
            <div className="form-group col-xs-5 col-sm-3 col-sm-offset-2">
              <button className="btn btn-block btn-primary" type="submit">Save</button>
            </div>
          </div>
        </form>
      </div>
    );
  },
});

export default HipchatConfig;
