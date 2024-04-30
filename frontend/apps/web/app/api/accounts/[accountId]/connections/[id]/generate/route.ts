import { withNeosyncContext } from '@/api-only/neosync-context';
import { Struct, Value } from '@bufbuild/protobuf';
import { GetAiGeneratedDataRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = GetAiGeneratedDataRequest.fromJson(await req.json());
    const response = await ctx.client.connectiondata.getAiGeneratedData(body);
    return {
      records: response.records.map((struct) => fromStructToRecord(struct)),
    };
  })(req);
}

function fromStructToRecord(struct: Struct): Record<string, unknown> {
  return Object.entries(struct.fields).reduce(
    (output, [key, value]) => {
      output[key] = handleValue(value);
      return output;
    },
    {} as Record<string, unknown>
  );
}

function handleValue(value: Value): unknown {
  switch (value.kind.case) {
    case 'structValue': {
      return fromStructToRecord(value.kind.value);
    }
    case 'listValue': {
      return value.kind.value.values.map((val) => handleValue(val));
    }
    default:
      return value.kind.value;
  }
}
