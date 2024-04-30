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

// eslint-disable-next-lint @typescript-eslint/no-explicit-any
function fromStructToRecord(struct: Struct): Record<string, any> {
  return Object.entries(struct.fields).reduce(
    (output, [key, value]) => {
      output[key] = handleValue(value);
      return output;
    },
    {} as Record<string, any>
  );
}

// eslint-disable-next-lint @typescript-eslint/no-explicit-any
function handleValue(value: Value): any {
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
