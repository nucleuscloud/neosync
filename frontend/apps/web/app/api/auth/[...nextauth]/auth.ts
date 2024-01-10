import { type TokenSet } from '@auth/core/types';
import { addSeconds, isAfter } from 'date-fns';
import NextAuth, { NextAuthConfig } from 'next-auth';

function getProviders(): NextAuthConfig['providers'] {
  const providers: NextAuthConfig['providers'] = [];
  const authConfig = getOAuthConfig();
  if (authConfig) {
    providers.push({
      id: authConfig.id,
      name: authConfig.name,
      type: authConfig.type,
      issuer: authConfig.expectedissuer,
      clientId: authConfig.clientId,
      clientSecret: authConfig.clientSecret ?? '',
      authorization: {
        url: authConfig.authorizeUrl,
        params: {
          audience: authConfig.audience,
          scope: authConfig.scope,
        },
      },
      userinfo: authConfig.userInfoUrl,
      token: authConfig.tokenUrl,

      wellKnown: getWellKnown(authConfig.issuer),
    });
  }

  return providers;
}

function getWellKnown(issuerUrl: string): string {
  return `${trimEnd(issuerUrl, '/')}/.well-known/openid-configuration`;
}

function trimEnd(val: string, chars: string): string {
  return val.endsWith(chars) ? val.substring(0, chars.length - 1) : val;
}

interface OAuthConfig {
  id: string;
  name: string;
  type: 'oidc';

  issuer: string;
  expectedissuer: string;

  authorizeUrl?: string;
  userInfoUrl?: string;
  tokenUrl?: string;

  clientId: string;
  clientSecret?: string;
  audience: string;
  scope: string;
}

function getOAuthConfig(): OAuthConfig | null {
  const issuer = process.env.AUTH_ISSUER;
  const clientId = process.env.AUTH_CLIENT_ID;
  const clientSecret = process.env.AUTH_CLIENT_SECRET;
  const audience = process.env.AUTH_AUDIENCE;
  const scope = process.env.AUTH_SCOPE;
  if (!issuer || !clientId || !audience || !scope) {
    return null;
  }

  const id = process.env.AUTH_PROVIDER_ID ?? 'unknown';
  const name = process.env.AUTH_PROVIDER_NAME ?? 'unknown';
  const expectedissuer = process.env.AUTH_EXPECTED_ISSUER || issuer;

  return {
    id,
    name,
    type: 'oidc',

    issuer,
    expectedissuer,
    clientId,
    clientSecret,
    audience,
    scope,

    authorizeUrl: process.env.AUTH_AUTHORIZE_URL,
    userInfoUrl: process.env.AUTH_USERINFO_URL,
    tokenUrl: process.env.AUTH_TOKEN_URL,
  };
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

        const oauthConfig = getOAuthConfig();
        if (!oauthConfig) {
          throw new Error('unable to find provider to refresh token');
        }

        const response = await fetch(await getTokenUrl(oauthConfig.issuer), {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: new URLSearchParams({
            client_id: oauthConfig.clientId,
            client_secret: oauthConfig.clientSecret ?? '',
            grant_type: 'refresh_token',
            refresh_token: (token as any).refreshToken, // eslint-disable-line @typescript-eslint/no-explicit-any
          }),
          method: 'POST',
        });
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
});

interface OidcConfiguration {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
}

async function getTokenUrl(issuer: string): Promise<string> {
  try {
    const oidcConfig = await getOpenIdConfiguration(issuer);
    if (!oidcConfig.token_endpoint) {
      throw new Error('unable to find token endpoint');
    }
    return oidcConfig.token_endpoint;
  } catch (err) {
    throw err;
  }
}

async function getOpenIdConfiguration(
  issuer: string
): Promise<Partial<OidcConfiguration>> {
  const wellKnownUrl = getWellKnown(issuer);

  const res = await fetch(wellKnownUrl, {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
  });
  return (await res.json()) as OidcConfiguration;
}

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
