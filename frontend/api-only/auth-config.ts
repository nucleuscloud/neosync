import { AuthOptions } from 'next-auth';
import Auth0 from 'next-auth/providers/auth0';
import { Provider } from 'next-auth/providers/index';

export const IS_AUTH_ENABLED = process.env.AUTH_ENABLED == 'true';

const providers: Provider[] = [];

if (IS_AUTH_ENABLED) {
  providers.push(
    Auth0({
      clientId: process.env.AUTH0_CLIENT_ID ?? '',
      clientSecret: process.env.AUTH0_CLIENT_SECRET ?? '',
      issuer: process.env.AUTH0_ISSUER,
      authorization: {
        params: {
          audience: process.env.AUTH0_AUDIENCE,
          scope: process.env.AUTH0_SCOPE,
        },
      },
    })
  );
}

export const authOptions: AuthOptions = {
  providers,
  session: { strategy: 'jwt' },
};
