'use client';
import { AiOutlineCloseCircle, AiOutlineWarning } from 'react-icons/ai';
import { SiTerraform } from 'react-icons/si';
import {
  Prism as SyntaxHighlighter,
  createElement,
} from 'react-syntax-highlighter';
interface Props {
  code: string;
}

export default function CodeBlock(props: Props) {
  const { code } = props;

  const custom = {
    property: {
      color: '#CACACA',
    },
    keyword: {
      color: '#6799D1',
    },
    boolean: {
      color: '#6799D1',
    },
    string: {
      color: '#B48872',
    },
    comment: {
      color: '#5A5D61',
    },
    'pre[class*="language-"]': {
      background: '#1d1f21',
      overflow: 'auto',
      borderRadius: '0.3em',
      tabSize: '4',
      padding: '1em',
      color: '#c5c8c6',
      textShadow: '0 1px rgba(0, 0, 0, 0.3)',
      wordSpacing: 'normal',
      lineHeight: '1.5',
    },
  };

  return (
    <div className="bg-[#1A1B1E] rounded-xl border border-gray-700">
      <div className="flex flex-row items-center pl-5 py-2">
        <SiTerraform className="text-purple-300" />
        <div className="text-gray-100 font-sm font-normal pl-2">
          terraform.tf
        </div>
      </div>
      <div className="overflow-hidden hidden lg:flex font-[14px] border-b border-b-gray-700 border-t border-t-gray-700">
        <SyntaxHighlighter
          language="hcl"
          showLineNumbers={true}
          style={custom}
          lineProps={{
            style: {
              wordBreak: 'break-all',
              whiteSpace: 'pre-wrap',
              fontSize: '14px',
            },
          }}
          wrapLines={false}
        >
          {code}
        </SyntaxHighlighter>
      </div>
      <div className="overflow-hidden flex lg:hidden font-[14px] border-b border-b-gray-700 border-t border-t-gray-700">
        <SyntaxHighlighter
          language="hcl"
          showLineNumbers={true}
          style={custom}
          renderer={({ rows, stylesheet, useInlineStyles }) => {
            return rows.map((row, index) => {
              const children = row.children;
              const lineNumberElement = children?.shift();

              /**
               * We will take current structure of the rows and rebuild it
               * according to the suggestion here https://github.com/react-syntax-highlighter/react-syntax-highlighter/issues/376#issuecomment-1246115899
               */
              if (lineNumberElement) {
                row.children = [
                  lineNumberElement,
                  {
                    children,
                    properties: {
                      className: [],
                    },
                    tagName: 'span',
                    type: 'element',
                  },
                ];
              }

              return createElement({
                node: row,
                stylesheet,
                useInlineStyles,
                key: index,
              });
            });
          }}
          lineProps={{
            style: {
              wordBreak: 'break-all',
              whiteSpace: 'pre-wrap',
              fontSize: '10px', //this is pretty small but the only size that will render code correctly on mobile
            },
          }}
          wrapLongLines={true}
        >
          {code}
        </SyntaxHighlighter>
      </div>
      <div className="flex flex-row items-center pl-2 py-2">
        <div className="flex flex-row items-center">
          <AiOutlineCloseCircle className="text-gray-500" />
          <div className="text-gray-500 pl-2">0</div>
        </div>
        <div className="flex flex-row items-center pl-5">
          <AiOutlineWarning className="text-gray-500" />
          <div className="text-gray-500 pl-2">0</div>
        </div>
      </div>
    </div>
  );
}
