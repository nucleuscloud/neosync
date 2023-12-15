import { withNeosyncContext } from '@/api-only/neosync-context';
import { NextRequest, NextResponse } from 'next/server';
import { getSystemAppConfig } from './config';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async () => getSystemAppConfig())(req);
}
