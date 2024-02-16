import { withNeosyncContext } from '@/api-only/neosync-context';
import { ValidateUserRegexCodeRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = ValidateUserRegexCodeRequest.fromJson(await req.json());
    return ctx.client.transformers.validateUserRegexCode(body);
  })(req);
}
