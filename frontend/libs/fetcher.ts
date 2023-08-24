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

export async function fetcher2<T = any>(
  url: string,
  info?: RequestInit
): Promise<T> {
  const res = await fetch(url, info);
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
