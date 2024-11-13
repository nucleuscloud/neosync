import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';
import protosidebar from './protos/proto-sidebars';

const protodocs = protosidebar.protodocs.map((item) => {
  if (item.type === 'category' && item.label === 'Files') {
    item.label = 'Protos';
  }
  return item;
});

const all = [
  { type: 'doc', label: 'Introduction', id: 'home' },
  { type: 'doc', label: 'Go', id: 'go' },
  { type: 'doc', label: 'TypeScript', id: 'typescript' },
  { type: 'doc', label: 'Python', id: 'python' },
  ...protodocs,
];

const sidebars: SidebarsConfig = {
  // By default, Docusaurus generates a sidebar from the docs folder structure

  protoSideBar: all as unknown as SidebarsConfig,
};

export default sidebars;
