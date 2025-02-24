'use client';

import React, { useState } from "react";
import styles from "./styles/Form.module.css";
import { client } from '@/lib/apollo-client';
import { ProvenanceClaimResponse, ProvenanceClaimVars } from '@/types/graphql';
import { GET_PROVENANCE_CLAIM } from '@/lib/queries';

const SearchByContentHashForm = ({ socket, toast }) => {
  const initialState = { contentHash: "" };
  const [formState, setFormState] = useState(initialState);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormState({ ...formState, [name]: value });
  };

  const queryContentHash = async (contentHash) => {
    const { data } = await client.query<ProvenanceClaimResponse, ProvenanceClaimVars>({
      query: GET_PROVENANCE_CLAIM,
      variables: { contentHash },
    });
    return data.provenanceClaims.items;
  }

  const submitForm = async (event) => {
    event.preventDefault();
    const { contentHash } = formState;
    if (contentHashisValid(contentHash) === false) {
      toast.error("Invalid Content Hash");
      return;
    }
try {
      const result = await queryContentHash(contentHash);
      toast.success(`Found ${result.length} match${result.length === 1 ? "" : "es"}`);
      if (result.length > 0) {
        // TODO: Navigate to the provenance claim page
      }
    } catch (error) {
      toast.error("Error querying content hash");
    }
  };

  const contentHashisValid = (hash) => {
    // blake3 hash is 32 bytes
    const hashRegex = /^0x[a-fA-F0-9]{64}$/;
    return hashRegex.test(hash);
  };

  const { contentHash } = formState;

  return (
    <form className={styles.Form} onSubmit={submitForm}>
      <div style={{ flexGrow: 1 }}>
        <div>Search by Content Hash</div>
        <input
          type="text"
          name="contentHash"
          id="contentHash"
          value={contentHash}
          placeholder="b3 hash"
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

export default SearchByContentHashForm;
