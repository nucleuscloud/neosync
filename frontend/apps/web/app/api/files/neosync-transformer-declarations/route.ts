import { withNeosyncContext } from '@/api-only/neosync-context';
import * as fs from 'fs';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async () => {
    const data = await fs.promises.readFile(
      '/app/apps/web/@types/neosync-transformers.d.ts',
      'utf8'
    );
    return data;
  })(req);
}
