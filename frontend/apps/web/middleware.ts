import { withAuth } from 'next-auth/middleware';
import { isAuthEnabled } from './api-only/auth-config';

// middleware is applied to all routes, use conditionals to select

export default withAuth(function middleware(_req) {}, {
  callbacks: {
    authorized: ({ token }) => {
      return isAuthEnabled() ? !!token : true;
    },
  },
});
