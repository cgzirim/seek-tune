import { ApolloClient, InMemoryCache } from '@apollo/client';

export const client = new ApolloClient({
  uri: 'https://indexer.royal.io/graphql',
  cache: new InMemoryCache(),
});