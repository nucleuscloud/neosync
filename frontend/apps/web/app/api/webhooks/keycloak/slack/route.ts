import { createHmac, timingSafeEqual } from 'crypto';
import { format } from 'date-fns';
import { utcToZonedTime } from 'date-fns-tz';
import got from 'got';
import { NextRequest, NextResponse } from 'next/server';
import * as Yup from 'yup';
import { Block } from './block';

const SIG_SECRET = process.env.KEYCLOAK_SLACK_WEBHOOK_HMAC_SECRET ?? 'foobar';
const KEYCLOAK_SIG_HEADER = 'X-Keycloak-Signature';
const SLACK_WEBHOOK_URL = process.env.KEYCLOAK_SLACK_WEBHOOK_URL;

const RegisterEvent = Yup.object({
  time: Yup.number().required(),
  type: Yup.string().oneOf(['access.REGISTER']).required(),
  authDetails: Yup.object({
    userId: Yup.string().required(),
    ipAddress: Yup.string().required(),
  }),
  details: Yup.object({
    email: Yup.string().required(),
    first_name: Yup.string(),
    last_name: Yup.string(),
  }).required(),
});
type RegisterEvent = Yup.InferType<typeof RegisterEvent>;

// Note, when testing this method, the body must be sent in the raw, unbeautified format for the signature to work correctly
export async function POST(req: NextRequest): Promise<NextResponse> {
  if (!SIG_SECRET) {
    return NextResponse.json(
      { message: 'missing signature secret in environment' },
      { status: 500 }
    );
  }
  if (!SLACK_WEBHOOK_URL) {
    return NextResponse.json(
      { message: 'missing slack webhook url' },
      { status: 500 }
    );
  }
  const incomingSignature = req.headers.get(KEYCLOAK_SIG_HEADER);
  if (!incomingSignature) {
    return NextResponse.json(
      { message: 'must provide sigure in header' },
      { status: 403 }
    );
  }

  try {
    const text = await req.text();

    const isTrusted = verifySignature(text, SIG_SECRET, incomingSignature);
    if (!isTrusted) {
      return NextResponse.json(
        {
          message:
            'the signature in the header differs from the computed request body',
        },
        { status: 403 }
      );
    }

    const registerEvent = await RegisterEvent.validate(JSON.parse(text));

    await got.post(SLACK_WEBHOOK_URL, {
      json: getSlackMessage(registerEvent),
    });

    return NextResponse.json({ message: 'ok', contents: JSON.parse(text) });
  } catch (err) {
    return NextResponse.json(
      { message: 'unable to complete request', error: err },
      { status: 500 }
    );
  }
}

function verifySignature(
  body: string,
  secret: string,
  untrustedSignature: string
): boolean {
  const signature = createHmac('sha256', secret).update(body).digest('hex');
  const trusted = Buffer.from(signature, 'ascii');
  const untrusted = Buffer.from(untrustedSignature, 'ascii');
  return timingSafeEqual(trusted, untrusted);
}

function getSlackMessage(event: RegisterEvent): { blocks: Block[] } {
  return {
    blocks: [
      {
        type: 'header',
        text: {
          type: 'plain_text',
          text: 'New Sign Up!',
        },
      },
      {
        type: 'section',
        fields: [
          {
            type: 'mrkdwn',
            text: `*IP*\n${event.authDetails.ipAddress}`,
          },
          {
            type: 'mrkdwn',
            text: `*User Id*\n${event.authDetails.userId}`,
          },
          {
            type: 'mrkdwn',
            text: `*Email*\n${event.details.email}`,
          },
          {
            type: 'mrkdwn',
            text: `*Name*\n${getFullname(event) ?? 'Unknown'}`,
          },
          {
            type: 'mrkdwn',
            text: `*When*\n${format(utcToZonedTime(new Date(event.time), 'America/Los_Angeles'), 'MMM d yyyy h:mma')}`,
          },
        ],
      },
    ],
  };
}

function getFullname(event: RegisterEvent): string | undefined {
  const pieces: string[] = [];

  if (event.details.first_name) {
    pieces.push(event.details.first_name);
  }
  if (event.details.last_name) {
    pieces.push(event.details.last_name);
  }
  return pieces.length > 0 ? pieces.join(' ') : undefined;
}
