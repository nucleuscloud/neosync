import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest, res: NextResponse) {
  const body = await req.json();

  const email = body.email;

  const formData = {
    legalConsentOptions: {
      consent: {
        consentToProcess: true,
        text: 'I agree to allow Example Company to store and process my personal data.',
        communications: [
          {
            value: true,
            subscriptionTypeId: 999,
            text: 'I agree to receive marketing communications from Example Company.',
          },
        ],
      },
    },
    fields: [
      {
        objectTypeId: '0-1',
        name: 'email',
        value: email,
      },
    ],
  };

  const options = {
    Authorization: `Bearer pat-na1-502b94c6-08ed-42b3-91f0-85ac0fbf9960`,
  };

  try {
    await axios.post(
      'https://api.hsforms.com/submissions/v3/integration/secure/submit/24034913/914cae02-64d1-469d-8b61-e1fbd22ab422',
      formData,
      { headers: options }
    );
    return NextResponse.json({ message: 'successfully create the lead' });
  } catch (e) {
    throw e;
  }
}
