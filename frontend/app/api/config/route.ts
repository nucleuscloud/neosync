import { isAuthEnabled } from '@/api-only/auth-config';
import { withNeosyncContext } from '@/api-only/neosync-context';
import { SystemAppConfig } from '@/app/config/app-config';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async () => {
    const sysConfig: SystemAppConfig = {
      isAuthEnabled: isAuthEnabled(),
      publicAppBaseUrl:
        process.env.NEXT_PUBLIC_APP_BASE_URL ?? 'http://localhost:3000',
    };
    return sysConfig;
  })(req);
}
