import { withNeosyncContext } from '@/api-only/neosync-context';
import * as fs from 'fs';
import { NextRequest, NextResponse } from 'next/server';
import path from 'path';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async () => {
    const filePath = path.resolve('@types/neosync-transformers.d.ts');
    const data = await fs.promises.readFile(filePath, 'utf8');
    return data;
  })(req);
}
