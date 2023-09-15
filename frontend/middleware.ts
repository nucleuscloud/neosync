import { withAuth } from 'next-auth/middleware';
import { IS_AUTH_ENABLED } from './api-only/auth-config';

// middleware is applied to all routes, use conditionals to select

export default withAuth(function middleware(_req) {}, {
  callbacks: {
    authorized: ({ token }) => {
      if (!IS_AUTH_ENABLED) {
        return true;
      }
      return !!token;
      // if (req.nextUrl.pathname.startsWith('/protected') && token === null) {
      //   return false;
      // }
      // return true;
    },
  },
});
