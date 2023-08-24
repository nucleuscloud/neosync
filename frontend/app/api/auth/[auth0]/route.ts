import { handleAuth } from '@auth0/nextjs-auth0';

export const GET = handleAuth({
  onError(_req: Request, error: Error) {
    console.error(error);
  },
});
