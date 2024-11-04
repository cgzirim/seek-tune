import { gql } from '@apollo/client';

export const GET_PROVENANCE_CLAIM = gql`
  query GetProvenanceClaim($contentHash: String!) {
    provenanceClaims(where: { contentHash: $contentHash }) {
      items {
        id
        originatorId
        originator {
          id
        }
        registrarId
        registrar {
          id
        }
        contentHash
        nftContract
        nftTokenId
        blockNumber
        transactionIndex
        createdAt
        updatedAt
        hourBucket
        dayBucket
        token {
          id
        }
        tokenDatabaseId
      }
    }
  }
`;