import { env } from '@/env';
import { generateStaticParams } from './blog/[...slug]/page';

const baseUrl = env.NEXT_PUBLIC_APP_URL;

export default async function sitemap() {
  const posts = await generateStaticParams(); //each slug gets resturned as an array with one string value

  const flattenedSlugs = posts.map((post) => post.slug[0]);

  const pages: any = [];

  flattenedSlugs.forEach((blog) => {
    const url = blog.replace('.md', '');
    pages.push({
      url: `${baseUrl}/blog/${url}`,
      lastModified: new Date().toISOString(),
    });
  });

  const routes = [
    '',
    '/about',
    '/docs',
    '/blog',
    '/solutions/fix-staging-environments',
    '/solutions/keep-environments-in-sync',
    '/solutions/security-privacy',
    '/solutions/unblock-local-development',
  ].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: new Date().toISOString(),
  }));

  return [...routes, ...pages];
}
