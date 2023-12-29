function shouldHighlightPathName(href: string, pathname: string): boolean {
  if (href === '/' && pathname === '/') {
    return true;
  }
  return href !== '/' && pathname.includes(href);
}

export function getPathNameHighlight(href: string, pathname: string): string {
  return shouldHighlightPathName(href, pathname)
    ? 'text-foreground'
    : 'text-foreground/60';
}
