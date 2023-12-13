import { type TokenSet } from '@auth/core/types';
import { addSeconds, isAfter } from 'date-fns';
import NextAuth from 'next-auth';

// console.log('auth0', getAuth0Provider());

function getAuth0Provider(): any {
  return {
    id: 'auth0',
    name: 'Auth0',
    type: 'oidc',
    issuer: process.env.AUTH0_ISSUER,
    clientId: process.env.AUTH0_CLIENT_ID ?? '',
    clientSecret: process.env.AUTH0_CLIENT_SECRET ?? '',
    authorization: {
      params: {
        audience: process.env.AUTH0_AUDIENCE,
        scope: process.env.AUTH0_SCOPE,
      },
    },
  };
}

export const {
  handlers: { GET, POST },
  auth,
} = NextAuth({
  providers: [
    {
      id: 'auth0',
      name: 'Auth0',
      type: 'oidc',
      issuer: process.env.AUTH0_ISSUER,
      clientId: process.env.AUTH0_CLIENT_ID ?? '',
      clientSecret: process.env.AUTH0_CLIENT_SECRET ?? '',
      authorization: {
        params: {
          audience: process.env.AUTH0_AUDIENCE,
          scope: process.env.AUTH0_SCOPE,
        },
      },
    },
  ],
  session: { strategy: 'jwt' },
  callbacks: {
    // authorized: async ({ auth }) => {
    //   console.log('hit authorized');
    //   return isAuthEnabled() ? !!auth : true;
    // },
    session: async ({ session, token }) => {
      (session as any).accessToken = (token as any).accessToken;
      return session;
    },
    jwt: async ({ token, account }) => {
      console.log('hit here');
      // Persist the OAuth access_token and or the user id to the token right after signin
      if (account) {
        token.accessToken = account.access_token;
        token.refreshToken = account.refresh_token;
        token.expiresAt = account.expires_at;
      }
      if (!token.expiresAt || isAfter(new Date(), (token as any).expiresAt)) {
        // refresh token
        if (!token.refreshToken) {
          // token can't be refreshed, fail
          throw new Error('session is expired, no refresh token available');
        }
        const auth0Provider = getAuth0Provider();
        const response = await fetch(
          `${auth0Provider.issuer ?? ''}/oauth/token`,
          {
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: new URLSearchParams({
              client_id: auth0Provider.clientId ?? '',
              client_secret: auth0Provider.clientSecret ?? '',
              grant_type: 'refresh_token',
              refresh_token: (token as any).refreshToken,
            }),
            method: 'POST',
          }
        );
        const tokens: TokenSet = await response.json();
        if (!response.ok) {
          throw tokens;
        }
        console.log('updating access token', tokens);
        token.accessToken = tokens.access_token;
        // the refresh token may not always be returned. If it's not, don't update
        if (tokens.refresh_token) {
          token.refreshToken = tokens.refresh_token;
        }
        if (tokens.expires_at) {
          token.expiresAt = tokens.expires_at;
        } else if (tokens.expires_in) {
          token.expiresAt = Math.floor(
            addSeconds(new Date(), tokens.expires_in).getTime() / 1000
          );
          console.log('updated token expires in', token.expiresAt);
        }
      }
      return token;
    },
  },
});
