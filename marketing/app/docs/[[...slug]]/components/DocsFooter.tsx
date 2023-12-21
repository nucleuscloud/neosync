import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { Doc } from 'contentlayer/generated';
import Link from 'next/link';
import { ReactElement } from 'react';

const githubBranch = 'main';
const githubBaseUrl = `https://github.com/nucleuscloud/neosync/marketing/blob/${githubBranch}/content/`;

interface Props {
  doc: Doc;
}

export default function DocsFooter({ doc }: Props): ReactElement {
  return (
    <>
      <hr className="my-8" />
      <div className="space-y-4 text-sm sm:flex sm:justify-between sm:space-y-0">
        <p className="m-0">
          Was this article helpful to you? <br />{' '}
          <Link
            className="inline-flex items-center space-x-1"
            target="_blank"
            rel="noreferrer"
            href="https://github.com/contentlayerdev/contentlayer/issues"
          >
            <span className="inline-block w-4">
              <GitHubLogoIcon />
            </span>
            <span>Provide feedback</span>
          </Link>
        </p>
        <p className="m-0 text-right">
          {/* Last edited on {format(new Date(doc.last_edited), 'MMMM dd, yyyy')}. */}
          <br />
          <Link
            className="inline-flex items-center space-x-1"
            target="_blank"
            rel="noreferrer"
            href={githubBaseUrl + doc._raw.sourceFilePath}
          >
            <span className="inline-block w-4">
              <GitHubLogoIcon />
            </span>
            <span>Edit this page</span>
          </Link>
        </p>
      </div>
    </>
  );
}
