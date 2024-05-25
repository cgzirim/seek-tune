import React, { useState } from "react";
import styles from "./styles/Form.module.css";

const Form = ({ socket }) => {
  const initialState = { spotifyUrl: "" };
  const [formState, setFormState] = useState(initialState);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormState({ ...formState, [name]: value });
  };

  const submitForm = (event) => {
    event.preventDefault();
    const { spotifyUrl } = formState;
    if (spotifyURLisValid(spotifyUrl) === false) {
      return;
    }

    socket.emit("newDownload", spotifyUrl);
    setFormState(initialState);
  };

  const spotifyURLisValid = (url) => {
    if (url.length === 0) {
      return false;
    }

    const splitURL = url.split("/");
    if (splitURL.length < 2) {
      return false;
    }

    return true;
  };

  const { spotifyUrl } = formState;

  return (
    <form className={styles.Form} onSubmit={submitForm}>
      <div style={{ flexGrow: 1 }}>
        <div>Add new songs</div>
        <input
          type="text"
          name="spotifyUrl"
          id="spotifyUrl"
          value={spotifyUrl}
          placeholder="https://open.spotify.com/.../..."
          onChange={handleChange}
        />
      </div>
      <input
        className={styles.Submit}
        type="submit"
        value="Submit"
        onClick={submitForm}
      />
    </form>
  );
};

export default Form;
