export function useGetInviteLink(token: string): string {
  return `${process.env.NEXT_PUBLIC_APP_BASE_URL}/invite?token=${token}`;
}
