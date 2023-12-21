const { createEnv } = require('@t3-oss/env-nextjs');
const { z } = require('zod');

function getNextPublicAppUrl() {
  if (process.env.NEXT_PUBLIC_APP_URL) {
    return process.env.NEXT_PUBLIC_APP_URL;
  }
  return `https://${process.env.VERCEL_URL}`;
}

const env = createEnv({
  server: {},
  client: {
    NEXT_PUBLIC_APP_URL: z.string().min(1),
  },
  runtimeEnv: {
    NEXT_PUBLIC_APP_URL: getNextPublicAppUrl(),
  },
});
module.exports = { env };
