export interface AskAiRequest {
  query: string;
  previousQueries?: string[];
}

export interface AskAiResponse {
  answer: AskAiAnswer;
}

export interface AskAiAnswer {
  text: string;
  pages: AskAiPage[];
  followupQuestions: string[];
}

export interface AskAiPage {
  space: string;
  revision: string;
  section: string[];
  page: string;
}

export interface GetLinkFromPageIdRequest {
  spaceId: string;
  pageId: string;
}

export interface GetLinkFromPageIdResponse {
  id: string;
  title: string;
  type: string;
  slug: string;
  path: string;
  description: string;
  pages: Page[];
}

type Page = (null | {
  type: 'link';
  href: string;
})[];

/**
 * Second parameter for an API Route.
 * ```ts
 * function GET(req: NextRequest, reqCtx: RequestContext): Promise<NextResponse>
 * ```
 */
export interface RequestContext {
  params: Record<string, string>;
}
