import mixpanel from 'mixpanel';
import { PropertyDict } from 'mixpanel';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest, res: NextResponse) {
  const Mixpanel = mixpanel.init('a970fed5baf076582713ccc5d63e09ca');

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
    await Mixpanel.track(name, props);
    return NextResponse.json({ message: 'success' });
  } catch (e) {
    throw e;
  }
}
