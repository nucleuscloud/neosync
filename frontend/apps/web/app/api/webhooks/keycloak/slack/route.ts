import { createHmac, timingSafeEqual } from 'crypto';
import { format } from 'date-fns';
import { toZonedTime } from 'date-fns-tz';
import { NextRequest, NextResponse } from 'next/server';
import * as Yup from 'yup';
import { Block } from './block';

const SIG_SECRET = process.env.KEYCLOAK_SLACK_WEBHOOK_HMAC_SECRET;
const KEYCLOAK_SIG_HEADER = 'X-Keycloak-Signature';
const SLACK_WEBHOOK_URL = process.env.KEYCLOAK_SLACK_WEBHOOK_URL;

/*
Example register event for username/password
{
  "time": 1711651627565,
  "realmId": "7a11ef96-d5ef-4bfe-beb8-a5d81f9de464",
  "uid": "d2e0fe26-f8e6-4012-b22f-87f989c3c9e7",
  "authDetails": {
    "realmId": "neosync-stage",
    "clientId": "neosync-app",
    "userId": "88251cc6-7408-4b31-a4c7-b2840d99b916",
    "ipAddress": "111.11.111.11",
    "username": "nick@example.com",
    "sessionId": "3cd8d025-d5d4-4ff2-82d3-a9c552ba422c"
  },
  "type": "access.REGISTER",
  "details": {
    "auth_method": "openid-connect",
    "auth_type": "code",
    "register_method": "form",
    "last_name": "Zelei",
    "redirect_uri": "https://app.stage.neosync.dev/api/auth/callback/neosync",
    "first_name": "Nick",
    "code_id": "3cd8d025-d5d4-4ff2-82d3-a9c552ba422c",
    "email": "nick@example.com",
    "username": "nick@example.com"
  }
}

Example register event for google
  {
    "time": 1713472787368,
    "realmId": "7a11ef96-d5ef-4bfe-beb8-a5d81f9de464",
    "uid": "cb444835-f5c3-4666-8a6f-1adfa6c0391d",
    "authDetails": {
      "realmId": "neosync-stage",
      "clientId": "neosync-app",
      "userId": "df684d95-fb16-457d-b461-3abfed5a7780",
      "ipAddress": "111.11.111.11",
      "username": "nickzelei@example.com",
      "sessionId": "6b4ef980-feff-418d-8e24-a2d6513b3f61"
    },
    "type": "access.REGISTER",
    "details": {
      "identity_provider": "google",
      "register_method": "broker",
      "identity_provider_identity": "nickzelei@example.com",
      "code_id": "6b4ef980-feff-418d-8e24-a2d6513b3f61",
      "email": "nickzelei@example.com",
      "username": "nickzelei@example.com"
      }
  }
*/
const RegisterEvent = Yup.object({
  time: Yup.number().required('The Time is required.'),
  type: Yup.string()
    .oneOf(['access.REGISTER'])
    .required('The Type is required.'),
  authDetails: Yup.object({
    userId: Yup.string().required('The userId is required.'),
    ipAddress: Yup.string().required('The IP Address is required.'),
  }),
  details: Yup.object({
    email: Yup.string().required('The Email is required.'),
    first_name: Yup.string(),
    last_name: Yup.string(),
    identity_provider: Yup.string(),
  }).required('The Details are required.'),
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

    await fetch(SLACK_WEBHOOK_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(getSlackMessage(registerEvent)),
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
  const trusted = new Uint8Array(Buffer.from(signature, 'ascii'));
  const untrusted = new Uint8Array(Buffer.from(untrustedSignature, 'ascii'));
  return timingSafeEqual(trusted, untrusted);
}

function getSlackMessage(event: RegisterEvent): { blocks: Block[] } {
  const detailsBlock: Block = {
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
        text: `*When*\n${format(toZonedTime(new Date(event.time), 'America/Los_Angeles'), 'MMM d yyyy h:mma')}`,
      },
    ],
  };

  const fullname = getFullname(event);
  if (fullname) {
    detailsBlock.fields.push({
      type: 'mrkdwn',
      text: `*Name*\n${fullname}`,
    });
  }

  if (event.details.identity_provider) {
    detailsBlock.fields.push({
      type: 'mrkdwn',
      text: `*IdP*\n${event.details.identity_provider}`,
    });
  }

  const blocks: Block[] = [
    {
      type: 'header',
      text: {
        type: 'plain_text',
        text: 'New Sign Up!',
      },
    },
    detailsBlock,
    {
      type: 'divider',
    },
  ];
  return {
    blocks: blocks,
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
