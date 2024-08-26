// eslint-disable-next-line @typescript-eslint/no-explicit-any
export async function fetcher<T = any>(url: string): Promise<T> {
  const res = await fetch(url, { credentials: 'include' });
  const body = await res.json();

  if (res.ok) {
    return body;
  }
  if (body.error) {
    throw new Error(body.error);
  }
  if (res.status > 399 && body.message) {
    throw new Error(body.message);
  }
  throw new Error('Unknown error when fetching');
}
