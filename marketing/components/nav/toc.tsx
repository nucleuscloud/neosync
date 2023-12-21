import { Nodes } from 'mdast';
import { toc } from 'mdast-util-toc';
import { remark } from 'remark';
import { Node } from 'unist';
import { visit } from 'unist-util-visit';

const textTypes = ['text', 'emphasis', 'strong', 'inlineCode'];

function flattenNode(node: Node) {
  const p: any[] = [];
  visit(node as any, (node) => {
    if (!textTypes.includes(node.type)) return;
    p.push((node as any).value);
  });
  return p.join(``);
}

interface Item {
  title: string;
  url: string;
  items?: Item[];
}

interface Items {
  items?: Item[];
}

function getItems(node: Nodes, current: Item): Items {
  if (!node) {
    return {};
  }

  if (node.type === 'paragraph') {
    visit(node as any, (item) => {
      if (item.type === 'link') {
        current.url = item.url;
        current.title = flattenNode(node);
      }

      if (item.type === 'text') {
        current.title = flattenNode(node);
      }
    });

    return current;
  }

  if (node.type === 'list') {
    const i = node.children.map((i) => getItems(i, { title: '', url: '' }));

    current.items = i as Item[];
    return current;
  } else if (node.type === 'listItem') {
    const heading = getItems(node.children[0], { title: '', url: '' });

    if (node.children.length > 1) {
      getItems(node.children[1], heading as Item);
    }

    return heading;
  }

  return {};
}

const getToc = () => (node: Nodes, file: any) => {
  const table = toc(node);
  if (table.map) {
    file.data = getItems(table.map, { title: '', url: '', items: [] });
  } else {
    file.data = {};
  }
};

export type TableOfContents = Items;

export async function getTableOfContents(
  content: string
): Promise<TableOfContents> {
  const result = await remark().use(getToc).process(content);

  return result.data;
}
