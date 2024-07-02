import { NextResponse } from 'next/server';
import { auth } from './app/api/auth/[...nextauth]/auth';
import { PUBLIC_PATHNAME, getSystemAppConfig } from './app/api/config/config';

export default auth((req) => {
  if (req.nextUrl.pathname.startsWith(PUBLIC_PATHNAME)) {
    const sysConfig = getSystemAppConfig();
    if (req.auth?.accessToken) {
      req.headers.set('Authorization', `Bearer ${req.auth.accessToken}`);
    }
    return NextResponse.rewrite(
      `${sysConfig.neosyncApiBaseUrl}${trimPrefix(req.nextUrl.pathname, PUBLIC_PATHNAME)}${req.nextUrl.search}`,
      {
        request: req,
      }
    );
  }
  return NextResponse.next();
});

function trimPrefix(str: string, prefix: string): string {
  if (str.startsWith(prefix)) {
    return str.slice(prefix.length);
  }
  return str;
}
