import { siteConfig } from '@/app/config/site';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import { cn } from '@/libs/utils';
import { GitHubLogoIcon, TwitterLogoIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import AccountSwitcher from './AccountSwitcher';
import { MainNav } from './MainNav';
import { MobileNav } from './MobileNav';
import { ModeToggle } from './ModeToggle';
import { UserNav } from './UserNav';
import { buttonVariants } from './ui/button';

export default function SiteHeader() {
  const authEnabled = useGetAuthEnabled();
  const iconButtonClassNames = cn(
    buttonVariants({
      variant: 'ghost',
    }),
    'p-0 px-2'
  );
  return (
    <header className="supports-backdrop-blur:bg-background/60 sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur">
      <div className="container flex h-14 items-center">
        <MainNav />
        <MobileNav />
        <div className="flex flex-1 md:justify-end">
          <nav className="flex items-center space-x-2">
            {authEnabled && <AccountSwitcher />}
            <Link
              href={siteConfig.links.github}
              target="_blank"
              rel="noreferrer"
            >
              <div className={iconButtonClassNames}>
                <GitHubLogoIcon />
                <span className="sr-only">GitHub</span>
              </div>
            </Link>
            <Link
              href={siteConfig.links.twitter}
              target="_blank"
              rel="noreferrer"
            >
              <div className={iconButtonClassNames}>
                <TwitterLogoIcon />
                <span className="sr-only">Twitter</span>
              </div>
            </Link>
            <ModeToggle />
            <UserNav />
          </nav>
        </div>
      </div>
    </header>
  );
}
