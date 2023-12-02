import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateUserDefinedTransformerRequest,
  DeleteUserDefinedTransformerRequest,
  GetUserDefinedTransformersRequest,
  UpdateUserDefinedTransformerRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.transformerClient.getUserDefinedTransformers(
      new GetUserDefinedTransformersRequest({ accountId })
    );
  })(req);
}

export async function DELETE(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const tId = searchParams.get('transformerId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.transformerClient.deleteUserDefinedTransformer(
      new DeleteUserDefinedTransformerRequest({ transformerId: tId })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateUserDefinedTransformerRequest.fromJson(await req.json());
    return ctx.transformerClient.createUserDefinedTransformer(body);
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateUserDefinedTransformerRequest.fromJson(await req.json());
    return ctx.transformerClient.updateUserDefinedTransformer(body);
  })(req);
}
