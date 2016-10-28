import React, {PropTypes} from 'react';

const AlertaConfig = React.createClass({
  propTypes: {
    config: PropTypes.shape({
      options: PropTypes.shape({
        environment: PropTypes.string,
        origin: PropTypes.string,
        token: PropTypes.bool,
        url: PropTypes.string,
      }).isRequired,
    }).isRequired,
    onSave: PropTypes.func.isRequired,
  },

  handleSaveAlert(e) {
    e.preventDefault();

    const properties = {
      environment: this.port.value,
      origin: this.from.value,
      token: this.username.value,
      url: this.password.value,
    };

    this.props.onSave(properties);
  },

  render() {
    const {environment, origin, token, url} = this.props.config.options;

    return (
      <div className="panel-body">
        <h4 className="text-center">Alerta Alert</h4>
        <br/>
        <form onSubmit={this.handleSaveAlert}>
          <div className="row">
            <div className="col-xs-7 col-sm-8 col-sm-offset-2">
              <p>
                Have alerts sent to Alerta
              </p>

              <div className="form-group">
                <label htmlFor="environment">Environment</label>
                <input className="form-control" id="environment" type="text" ref={(r) => this.environment = r} defaultValue={environment || ''}></input>
              </div>

              <div className="form-group">
                <label htmlFor="origin">Origin</label>
                <input className="form-control" id="origin" type="text" ref={(r) => this.origin = r} defaultValue={origin || ''}></input>
              </div>

              <div className="form-group">
                <label htmlFor="token">Token</label>
                <input className="form-control" id="token" type="text" ref={(r) => this.token = r} defaultValue={token || ''}></input>
                <span>Note: a value of <code>true</code> indicates the Alerta Token has been set</span>
              </div>

              <div className="form-group">
                <label htmlFor="url">User</label>
                <input className="form-control" id="url" type="text" ref={(r) => this.url = r} defaultValue={url || ''}></input>
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

export default AlertaConfig;
