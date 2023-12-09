import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import {
  CreateUserDefinedTransformerRequest,
  DeleteUserDefinedTransformerRequest,
  GetUserDefinedTransformersRequest,
  UpdateUserDefinedTransformerRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.transformers.getUserDefinedTransformers(
      new GetUserDefinedTransformersRequest({ accountId: params.accountId })
    );
  })(req);
}

export async function DELETE(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const tId = searchParams.get('transformerId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.client.transformers.deleteUserDefinedTransformer(
      new DeleteUserDefinedTransformerRequest({ transformerId: tId })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateUserDefinedTransformerRequest.fromJson(await req.json());
    return ctx.client.transformers.createUserDefinedTransformer(body);
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateUserDefinedTransformerRequest.fromJson(await req.json());
    return ctx.client.transformers.updateUserDefinedTransformer(body);
  })(req);
}
