export function getMobileMainNav(
  accountName: string
): { title: string; href: string }[] {
  return [
    // {
    //   title: 'Overview',
    //   href: `/`,
    // },
    {
      title: 'Jobs',
      href: `/${accountName}/jobs`,
    },
    {
      title: 'Runs',
      href: `/${accountName}/runs`,
    },
    {
      title: 'Transformers',
      href: `/${accountName}/transformers`,
    },
    {
      title: 'Connections',
      href: `/${accountName}/connections`,
    },
    {
      title: 'Settings',
      href: `/${accountName}/settings`,
    },
  ];
}
