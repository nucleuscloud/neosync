'use client';
import { type DocHeading } from '@/contentlayer.config';
import { ChevronRightIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useState } from 'react';
import { getNodeText, sluggifyTitle } from './sluggify';

interface Props {
  headings: DocHeading[];
}

export default function PageNavigation({
  headings,
}: Props): ReactElement | null {
  const [activeHeading, setActiveHeading] = useState('');

  useEffect(() => {
    const handleScroll = () => {
      let current = '';
      for (const heading of headings) {
        const slug = sluggifyTitle(getNodeText(heading.title));
        const element = document.getElementById(slug);
        if (element && element.getBoundingClientRect().top < 240)
          current = slug;
      }
      setActiveHeading(current);
    };
    handleScroll();
    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  }, [headings]);

  const headingsToRender = headings.filter((_) => _.level > 1);

  if ((headingsToRender ?? []).length === 0) return null;

  return (
    <div className="text-sm">
      <h4 className="mb-4 font-medium text-slate-600 dark:text-slate-300">
        On this page
      </h4>
      <ul className="space-y-2">
        {headingsToRender.map(({ title, level }, index) => (
          <li key={index}>
            <a
              href={`#${sluggifyTitle(getNodeText(title))}`}
              style={{ marginLeft: (level - 2) * 16 }}
              className={`flex ${
                sluggifyTitle(getNodeText(title)) == activeHeading
                  ? 'text-violet-600 dark:text-violet-400'
                  : 'hover:text-slate-600 dark:hover:text-slate-300'
              }`}
            >
              <span className="mr-2 mt-[5px] block w-1.5 shrink-0">
                <ChevronRightIcon />
              </span>
              <span
                dangerouslySetInnerHTML={{
                  __html: title
                    .replace('`', '<code style="font-size: 0.75rem;">')
                    .replace('`', '</code>'),
                }}
              />
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
