import { type TokenSet } from '@auth/core/types';
import { addSeconds, isAfter } from 'date-fns';
import NextAuth, { NextAuthConfig } from 'next-auth';

function getProviders(): NextAuthConfig['providers'] {
  const providers: NextAuthConfig['providers'] = [];

  const auth0Config = getAuth0Config();
  if (auth0Config) {
    providers.push({
      id: 'auth0',
      name: 'Auth0',
      type: 'oidc',
      issuer: auth0Config.issuer,
      clientId: auth0Config.clientId,
      clientSecret: auth0Config.clientSecret,
      authorization: {
        params: {
          audience: auth0Config.audience,
          scope: auth0Config.scope,
        },
      },
      wellKnown: `${auth0Config.issuer}/.well-known/openid-configuration`,
    });
  }

  return providers;
}

function getAuth0Config(): Auth0Config | null {
  const issuer = process.env.AUTH0_ISSUER;
  const clientId = process.env.AUTH0_CLIENT_ID;
  const clientSecret = process.env.AUTH0_CLIENT_SECRET;
  const audience = process.env.AUTH0_AUDIENCE;
  const scope = process.env.AUTH0_SCOPE;

  if (!issuer || !clientId || !clientSecret || !audience || !scope) {
    return null;
  }
  return {
    issuer,
    clientId,
    clientSecret,
    audience,
    scope,
  };
}

interface Auth0Config {
  issuer: string;
  clientId: string;
  clientSecret: string;
  audience: string;
  scope: string;
}

export const {
  handlers: { GET, POST },
  // auth function meant to be used in RSC or middleware.
  auth,
  // server-side signIn. Use signIn from the next-auth/react library for client-side
  signIn,
  // server-side signOut. Use signOut from the next-auth/react library for client-side
  signOut,
} = NextAuth({
  providers: getProviders(),
  session: { strategy: 'jwt' },
  callbacks: {
    session: async ({ session, token }) => {
      session.accessToken = (token as any).accessToken; // eslint-disable-line @typescript-eslint/no-explicit-any
      return session;
    },
    jwt: async ({ token, account }) => {
      // Persist the OAuth access_token and or the user id to the token right after signin
      if (account) {
        token.accessToken = account.access_token;
        token.refreshToken = account.refresh_token;
        token.expiresAt = account.expires_at;
        token.provider = account.provider;
      }
      if (
        !token.expiresAt ||
        // Both times must be in the same format
        isAfter(new Date(), new Date((token as any).expiresAt * 1000)) // eslint-disable-line @typescript-eslint/no-explicit-any
      ) {
        // refresh token
        if (!token.refreshToken) {
          // token can't be refreshed, fail
          throw new Error('session is expired, no refresh token available');
        }
        if (!token.provider) {
          throw new Error(
            'unable to find provider to be used to refresh token'
          );
        }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        switch ((token as any).provider) {
          case 'auth0': {
            const auth0Provider = getAuth0Config();
            if (!auth0Provider) {
              throw new Error('unable to find auth0 provider to refresh token');
            }
            const response = await fetch(
              `${auth0Provider?.issuer ?? ''}/oauth/token`,
              {
                headers: {
                  'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                  client_id: auth0Provider?.clientId ?? '',
                  client_secret: auth0Provider?.clientSecret ?? '',
                  grant_type: 'refresh_token',
                  refresh_token: (token as any).refreshToken, // eslint-disable-line @typescript-eslint/no-explicit-any
                }),
                method: 'POST',
              }
            );
            const tokens: TokenSet = await response.json();
            if (!response.ok) {
              throw tokens;
            }
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
            }
            break;
          }
          default: {
            throw new Error(
              'unknown or unsupported provider type, unable to refresh token'
            );
          }
        }
      }
      return token;
    },
  },
});

declare module 'next-auth' {
  export interface Session {
    accessToken: string;
  }
}

// This isn't currently working, guessing because next-auth has its own local dependency of @auth/core due to mismatched versions
// declare module '@auth/core/jwt' {
//   export interface JWT {
//     refreshToken?: string;
//     expiresAt?: number;
//     accessToken?: string;
//   }
// }
