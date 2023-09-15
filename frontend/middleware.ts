import { withAuth } from 'next-auth/middleware';

// middleware is applied to all routes, use conditionals to select

export default withAuth(function middleware(_req) {}, {
  callbacks: {
    authorized: ({ token }) => {
      return !!token;
      // if (req.nextUrl.pathname.startsWith('/protected') && token === null) {
      //   return false;
      // }
      // return true;
    },
  },
});
