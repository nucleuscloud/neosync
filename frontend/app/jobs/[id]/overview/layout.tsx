'use client';
import { SidebarNav } from '@/components/SideBarNav';
import { Separator } from '@/components/ui/separator';
import { useParams } from 'next/navigation';

interface SettingsLayoutProps {
  children: React.ReactNode;
}

export default function SettingsLayout({ children }: SettingsLayoutProps) {
  const params = useParams();
  const basePath = `/jobs/${params.id}`;

  const sidebarNavItems = [
    {
      title: 'Overview',
      href: `${basePath}/overview`,
    },
    {
      title: 'Source',
      href: `${basePath}/overview/source`,
    },
    {
      title: 'Destinations',
      href: `${basePath}/overview/destinations`,
    },
    {
      title: 'Schema',
      href: `${basePath}/overview/schema`,
    },
  ];

  return (
    <div className="hidden space-y-6 p-10 pb-16 md:block">
      <div className="space-y-0.5">
        <h2 className="text-2xl font-bold tracking-tight">Job Settings</h2>
        <p className="text-muted-foreground">Manage your job settings.</p>
      </div>
      <Separator className="my-6" />
      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        <aside className="-mx-4 lg:w-1/5">
          <SidebarNav items={sidebarNavItems} />
        </aside>
        <div className="flex-1 lg:max-w-8xl">{children}</div>
      </div>
    </div>
  );
}
