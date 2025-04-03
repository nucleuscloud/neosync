import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ReactElement, useMemo } from 'react';

interface Props {
  reporterTypes: string[];
}

export default function ReporterTypeCell(props: Props): ReactElement {
  const { reporterTypes } = props;

  const uniqueReporterTypes = dedupe(reporterTypes);

  return (
    <span className="max-w-[500px] truncate font-medium">
      <div className="flex flex-col lg:flex-row items-start gap-1">
        {uniqueReporterTypes.map((reporterType) => (
          <ReporterTypeBadge key={reporterType} reporterType={reporterType} />
        ))}
      </div>
    </span>
  );
}

function dedupe(arr: string[]): string[] {
  return [...Array.from(new Set(arr))];
}

interface ReporterTypeBadgeProps {
  reporterType: string;
}

function ReporterTypeBadge(props: ReporterTypeBadgeProps): ReactElement {
  const { reporterType } = props;
  const tooltip = useReporterTypeTooltip(reporterType);
  return (
    <TooltipProvider>
      <Tooltip delayDuration={200}>
        <TooltipTrigger type="button">
          <Badge
            variant="outline"
            className="text-xs bg-blue-100 text-gray-800 cursor-default dark:bg-blue-200 dark:text-gray-900"
          >
            {reporterType}
          </Badge>
        </TooltipTrigger>
        <TooltipContent>{tooltip}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function useReporterTypeTooltip(reporterType: string): string {
  return useMemo(() => {
    switch (reporterType) {
      case 'regex':
        return 'Classified using regular expressions';
      case 'llm':
        return 'Classified using an LLM';
      default:
        return reporterType;
    }
  }, [reporterType]);
}
