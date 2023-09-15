import { AuthOptions } from 'next-auth';
import Auth0 from 'next-auth/providers/auth0';

export const authOptions: AuthOptions = {
  providers: [
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
    }),
  ],
  session: { strategy: 'jwt' },
};
