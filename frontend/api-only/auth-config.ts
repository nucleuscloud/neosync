import { AuthOptions } from 'next-auth';
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
  };
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
