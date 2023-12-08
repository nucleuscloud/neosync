import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';
import protosidebar from './protos/proto-sidebars';

const protodocs =
  protosidebar.protodocs.find(
    (item) => item.type === 'category' && item.label === 'Files'
  )?.items ?? [];

const all = [{ type: 'doc', label: 'Overview', id: 'home' }, ...protodocs];

const sidebars: SidebarsConfig = {
  // By default, Docusaurus generates a sidebar from the docs folder structure

  protoSideBar: all as unknown as SidebarsConfig,
};

export default sidebars;
