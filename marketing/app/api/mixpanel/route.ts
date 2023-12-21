import { env } from '@/env';
import mixpanel, { PropertyDict } from 'mixpanel';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest) {
  const mixpanelToken = env.MIXPANEL_TOKEN;
  if (!mixpanelToken) {
    return NextResponse.json({ message: 'no token' });
  }
  const Mixpanel = mixpanel.init(mixpanelToken);

  const body = await req.json();

  const name: string = body.name ?? '';
  const props: PropertyDict = body.props ?? {};

  var data: PropertyDict = { distinct_id: name };
  for (const key in data) {
    if (data.hasOwnProperty(key)) {
      props[key] = data[key];
    }
  }

  try {
    Mixpanel.track(name, props);
    return NextResponse.json({ message: 'success' });
  } catch (e) {
    throw e;
  }
}
