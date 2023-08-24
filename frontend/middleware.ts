import { withMiddlewareAuthRequired } from '@auth0/nextjs-auth0/edge';
import { NextResponse } from 'next/server';

export default withMiddlewareAuthRequired(async function middleware(_req) {
  const res = NextResponse.next();
  // custom middleware here
  return res;
});

// export const config = {};
