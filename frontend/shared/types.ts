/**
 * Second parameter for an API Route.
 * ```ts
 * function GET(req: NextRequest, reqCtx: RequestContext): Promise<NextResponse>
 * ```
 */
export interface RequestContext {
  params: Record<string, string>;
}
