import React from "react";
import { render } from "react-dom";

class LocustApp extends React.PureComponent {
  state = {
    target: "",
    results: {},
  };

  render() {
    return (
      <div>
        <h3>Where we droppin fam?</h3>
        <input
          placeholder="https://google.com"
          value={this.state.target}
          onChange={e => this.setState({ target: e.target.value })}
        />
        <button onClick={this._swarm}>Let's Get that Bread</button>
        <hr />
        {Object.keys(this.state.results).map(key => (
          <div>
            <b>{key}:</b>
            <span>{this.state.results[key]}</span>
          </div>
        ))}
      </div>
    );
  }

  _swarm = () => {
    const { target } = this.state;
    if (target === "") {
      return;
    }

    fetch(`/request?target=${encodeURI(target)}`)
      .then(res => res.text())
      .then(this._pollResults);
  };

  _pollResults = requestID => {
    fetch(`/results?requestID=${requestID}`)
      .then(response => {
        var text = "";
        var reader = response.body.getReader();
        var decoder = new TextDecoder();


        const readChunk = () => {
          return reader.read().then(appendChunks);
        }

        const appendChunks = (result) => {
          const chunk = decoder.decode(result.value || new Uint8Array(), {
            stream: !result.done
          });
          const results = JSON.parse(chunk);
          this.setState({results});
          if (result.done) {return 'done';}
          return readChunk();
        }

        return readChunk();
      })
      .then(() => {
        alert('All done \u1F975')
      })

    function onChunkedResponseComplete(result) {
      console.log("all done!", result);
    }

    function onChunkedResponseError(err) {
      console.error(err);
    }

    function processChunkedResponse(response) {
    }
  };
}

render(React.createElement(LocustApp), document.getElementById("main"));
