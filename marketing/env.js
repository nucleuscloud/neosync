const { createEnv } = require('@t3-oss/env-nextjs');
const { z } = require('zod');

function getNextPublicAppUrl() {
  if (process.env.NEXT_PUBLIC_APP_URL) {
    return process.env.NEXT_PUBLIC_APP_URL;
  }
  return `https://${process.env.VERCEL_URL}`;
}

const env = createEnv({
  server: {
    MIXPANEL_TOKEN: z.string().optional(),
    LOOPS_FORM_ID: z.string().optional(),
  },
  client: {
    NEXT_PUBLIC_APP_URL: z.string().min(1),
  },
  runtimeEnv: {
    NEXT_PUBLIC_APP_URL: getNextPublicAppUrl(),
    MIXPANEL_TOKEN: process.env.MIXPANEL_TOKEN,
    HUBSPOT_TOKEN: process.env.HUBSPOT_TOKEN,
    LOOPS_FORM_ID: process.env.LOOPS_FORM_ID,
  },
});
module.exports = { env };
