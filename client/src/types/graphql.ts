export interface ProvenanceClaimResponse {
    provenanceClaims: {
      items: {
        id: string;
        originatorId: string;
        originator: {
          id: string;
        };
        registrarId: string;
        registrar: {
          id: string;
        };
        contentHash: string;
        nftContract: string;
        nftTokenId: string;
        blockNumber: string;
        transactionIndex: number;
        createdAt: string;
        updatedAt: string;
        hourBucket: number;
        dayBucket: number;
        token?: {
          id: string;
        };
        tokenDatabaseId?: string;
      }[];
    };
  }
  
  export interface ProvenanceClaimVars {
    contentHash: string;
  }