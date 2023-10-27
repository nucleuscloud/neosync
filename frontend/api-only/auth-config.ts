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
  }

  return {
    providers,
    session: { strategy: 'jwt' },
    callbacks: {
      async jwt({ token, account }) {
        // Persist the OAuth access_token and or the user id to the token right after signin
        if (account) {
          token.accessToken = account.access_token;
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
