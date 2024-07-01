import { NextRequest, NextResponse } from 'next/server';
import { auth } from './app/api/auth/[...nextauth]/auth';
import { PUBLIC_PATHNAME, getSystemAppConfig } from './app/api/config/config';

const middleware = auth(function middleware(request: NextRequest) {
  if (request.nextUrl.pathname.startsWith(PUBLIC_PATHNAME)) {
    const sysConfig = getSystemAppConfig();
    return NextResponse.rewrite(
      `${sysConfig.neosyncApiBaseUrl}${trimPrefix(request.nextUrl.pathname, PUBLIC_PATHNAME)}${request.nextUrl.search}`
    );
  }
  return NextResponse.next();
});
export default middleware;

function trimPrefix(str: string, prefix: string): string {
  if (str.startsWith(prefix)) {
    return str.slice(prefix.length);
  }
  return str;
}
