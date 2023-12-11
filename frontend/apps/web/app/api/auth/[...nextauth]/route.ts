import { getAuthOptions } from '@/api-only/auth-config';
import NextAuth from 'next-auth';

const handler = NextAuth(getAuthOptions());

export { handler as GET, handler as POST };
