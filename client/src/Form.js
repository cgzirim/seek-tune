import React, { Component } from "react";

class Form extends Component {
  initialState = { spotifyUrl: "" };

  state = this.initialState;

  handleChange = (event) => {
    const { name, value } = event.target;

    this.setState({ [name]: value });
  };

  submitForm = () => {
    const { spotifyUrl } = this.state;
    const { socket } = this.props;
    if (this.spotifyURLisValid(spotifyUrl) === false) {
      return;
    }

    socket.emit("newDownload", spotifyUrl);
    console.log("newDownload: ", spotifyUrl);
    this.setState(this.initialState);
  };

  spotifyURLisValid = (url) => {
    if (url.length === 0) {
      console.log("Spotify URL required");
      return false;
    }

    const splitURL = url.split("/");
    if (splitURL.length < 2) {
      console.log("Invalid Spotify URL format");
      return false;
    }

    let spotifyID = splitURL[splitURL.length - 1];
    if (spotifyID.includes("?")) {
      spotifyID = spotifyID.split("?")[0];
    }

    // Check if the Spotify ID is alphanumeric
    if (!/^[a-zA-Z0-9]+$/.test(spotifyID)) {
      console.log("Invalid Spotify ID format");
      return false;
    }

    // Check if the Spotify ID is of expected length
    if (spotifyID.length !== 22) {
      console.log("Invalid Spotify ID length");
      return false;
    }

    // Additional validation logic can be added here

    return true;
  };

  render() {
    const { spotifyUrl } = this.state;

    return (
      <form>
        <label htmlFor="spotifyUrl">spotifyUrl</label>
        <input
          type="text"
          name="spotifyUrl"
          id="spotifyUrl"
          value={spotifyUrl}
          placeholder="https://open.spotify.com/.../..."
          onChange={this.handleChange}
        />
        <input type="button" value="Submit" onClick={this.submitForm} />
      </form>
    );
  }
}

export default Form;
