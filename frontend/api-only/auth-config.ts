import { type TokenSet } from '@auth/core/types';
import { addSeconds, isAfter } from 'date-fns';
import { AuthOptions } from 'next-auth';
import { JWT } from 'next-auth/jwt';
import Auth0, { Auth0Profile } from 'next-auth/providers/auth0';
import { OAuthConfig, Provider } from 'next-auth/providers/index';

export function isAuthEnabled(): boolean {
  return process.env.AUTH_ENABLED == 'true';
}

export function getAuthOptions(): AuthOptions {
  const providers: Provider[] = [];

  if (isAuthEnabled()) {
    providers.push(getAuth0Provider());
    console.log('auth0', getAuth0Provider());
  }

  return {
    providers,
    session: { strategy: 'jwt' },
    callbacks: {
      async jwt({ token, account }) {
        // Persist the OAuth access_token and or the user id to the token right after signin
        if (account) {
          token.accessToken = account.access_token;
          token.refreshToken = account.refresh_token;
          token.expiresAt = account.expires_at;
        }
        if (!!token.expiresAt && isAfter(new Date(), token.expiresAt)) {
          // refresh token
          if (!token.refreshToken) {
            // token can't be refreshed, fail
            throw new Error('session is expired, no refresh token available');
          }
          const auth0Provider = getAuth0Provider();
          const response = await fetch(
            `${auth0Provider.options?.issuer ?? ''}/oauth/token`,
            {
              headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
              body: new URLSearchParams({
                client_id: auth0Provider.options?.clientId ?? '',
                client_secret: auth0Provider.options?.clientSecret ?? '',
                grant_type: 'refresh_token',
                refresh_token: token.refreshToken,
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
        }
        return token;
      },
    },
  };
}

/**
 * Module augmentation for `next-auth` types. Allows us to add custom properties to the `session`
 * object and keep type safety.
 *
 * @see https://next-auth.js.org/getting-started/typescript#module-augmentation
 */
declare module 'next-auth/jwt' {
  interface JWT {
    accessToken?: string;
    expiresAt?: number;
    refreshToken?: string;
  }
}

function getAuth0Provider(): OAuthConfig<Auth0Profile> {
  return Auth0({
    clientId: process.env.AUTH0_CLIENT_ID ?? '',
    clientSecret: process.env.AUTH0_CLIENT_SECRET ?? '',
    issuer: process.env.AUTH0_ISSUER,
    authorization: {
      params: {
        audience: process.env.AUTH0_AUDIENCE,
        scope: process.env.AUTH0_SCOPE,
      },
    },
  });
}
