// import { format, fromUnixTime } from 'date-fns';
// import { utcToZonedTime } from 'date-fns-tz';
// // import { getServerSession } from 'next-auth';
// import { NextRequest, NextResponse } from 'next/server';
// // import { getAuthOptions } from '../../../../api-only/auth-config';

import { NextRequest, NextResponse } from 'next/server';
import { auth } from '../../auth/[...nextauth]/auth';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const res = NextResponse.next();
  const session = await auth(req as any, res as any);
  console.log(session);
  // const session = await getServerSession(getAuthOptions());
  // console.log('GET session', session);
  // const jwt = await getToken({ req });
  // const jwt = session;

  return res;
}
