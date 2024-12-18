import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import SwitchCard from '@/components/switches/SwitchCard';
import { Button } from '@/components/ui/button';
import { fromJson } from '@bufbuild/protobuf';
import { JobMapping, JobMappingSchema } from '@neosync/sdk';
import { filesize } from 'filesize';
import {
  ReactElement,
  SetStateAction,
  useCallback,
  useEffect,
  useState,
} from 'react';
import { toast } from 'sonner';
import { FileProcessingResult, FileUpload } from '../../FileUpload/FileUpload';

export interface ImportMappingsConfig {
  truncateAll: boolean;
  overrideOverlapping: boolean;
}

interface Props {
  onImport(jobmappings: JobMapping[], config: ImportMappingsConfig): void;
}

interface ExtractedJobMappings {
  lastModified: number;
  mappings: JobMapping[];
}

export default function ImportJobMappingsButton(props: Props): ReactElement {
  const { onImport } = props;
  const [jmExtracted, setJmExtracted] = useState<
    Record<string, ExtractedJobMappings>
  >({});
  const [truncateAll, setTruncateAll] = useState(false);
  const [overrideOverlapping, setOverrideOverlapping] = useState(false);
  return (
    <div>
      <ConfirmationDialog
        trigger={
          <Button type="button" variant="outline">
            <ButtonText text="Import" />
          </Button>
        }
        headerText="Import Job Mappings"
        description="Multiple files may be uploaded."
        body={
          <Body
            setJobMappings={setJmExtracted}
            overrideOverlapping={overrideOverlapping}
            setOverrideOverlapping={setOverrideOverlapping}
            truncateAll={truncateAll}
            setTruncateAll={setTruncateAll}
          />
        }
        containerClassName="max-w-xl"
        buttonText="Import"
        onConfirm={() => {
          onImport(
            squashJobMappings(getSortedJobMappingsFromExtracted(jmExtracted)),
            {
              overrideOverlapping: overrideOverlapping,
              truncateAll: truncateAll,
            }
          );
          setJmExtracted({});
          setTruncateAll(false);
          setOverrideOverlapping(false);
        }}
      />
    </div>
  );
}

function getSortedJobMappingsFromExtracted(
  input: Record<string, ExtractedJobMappings>
): JobMapping[][] {
  // Sort the entires by last modified date oldest first
  const sortedEntries = Object.entries(input).sort(
    ([, a], [, b]) => a.lastModified - b.lastModified
  );
  return sortedEntries.map(([, values]) => values.mappings);
}

function squashJobMappings(input: JobMapping[][]): JobMapping[] {
  const seen = new Set<string>();
  return input.reduce((output: JobMapping[], curr: JobMapping[]) => {
    curr.forEach((mapping) => {
      const key = `${mapping.schema}.${mapping.table}.${mapping.column}`;
      if (!seen.has(key)) {
        seen.add(key);
        output.push(mapping);
      }
    });

    return output;
  }, []);
}

// Async Worker code that handles parsing the input data as JSON
const workerCode = `
  self.onmessage = function(e) {
    const { file, id } = e.data;
    const reader = new FileReader();

    reader.onload = function(event) {
      try {
        const json = JSON.parse(event.target.result);
        self.postMessage({
          id,
          success: true,
          data: json,
          fileName: file.name
        });
      } catch (error) {
        self.postMessage({
          id,
          success: false,
          error: 'Invalid JSON format',
          fileName: file.name
        });
      }
    };

    reader.onerror = function() {
      self.postMessage({
        id,
        success: false,
        error: 'Error reading file',
        fileName: file.name
      });
    };

    reader.readAsText(file);
  };
`;

const workerBlob = new Blob([workerCode], { type: 'application/javascript' });

interface BodyProps {
  setJobMappings(
    method: SetStateAction<Record<string, ExtractedJobMappings>>
  ): void;

  overrideOverlapping: boolean;
  setOverrideOverlapping(value: boolean): void;
  truncateAll: boolean;
  setTruncateAll(value: boolean): void;
}

function Body(props: BodyProps): ReactElement {
  const {
    setJobMappings,
    overrideOverlapping,
    setOverrideOverlapping,
    truncateAll,
    setTruncateAll,
  } = props;

  const [worker, setWorker] = useState<Worker | null>(null);

  useEffect(() => {
    const workerUrl = URL.createObjectURL(workerBlob);
    const newWorker = new Worker(workerUrl);
    setWorker(newWorker);

    return () => {
      newWorker.terminate();
      URL.revokeObjectURL(workerUrl);
    };
  }, []);

  const processFile = useCallback(
    async (file: File): Promise<FileProcessingResult<JobMapping[]>> => {
      return new Promise((resolve) => {
        if (!worker) {
          resolve({
            success: false,
            error: 'Worker not initialized',
          });
          return;
        }

        const processId = Date.now().toString() + Math.random().toString(36);

        const handler = (e: MessageEvent) => {
          const { id, success, error, data } = e.data;
          if (id === processId) {
            worker.removeEventListener('message', handler);

            if (!success) {
              resolve({
                success: false,
                error: error,
              });
              return;
            }

            try {
              if (Array.isArray(data)) {
                const mappings = data.map((d) => fromJson(JobMappingSchema, d));
                resolve({
                  success: true,
                  data: mappings,
                });
              } else {
                resolve({
                  success: false,
                  error: 'Invalid data format',
                });
              }
            } catch (err) {
              resolve({
                success: false,
                error: err instanceof Error ? err.message : 'Processing failed',
              });
            }
          }
        };

        worker.addEventListener('message', handler);
        worker.postMessage({ file, id: processId });
      });
    },
    [worker]
  );
  const handleSuccess = useCallback(
    (file: File, mappings: JobMapping[]) => {
      const newVal: ExtractedJobMappings = {
        mappings,
        lastModified: file.lastModified ?? 0,
      };
      setJobMappings((prev) => ({
        ...prev,
        [file.name]: newVal,
      }));
    },
    [setJobMappings]
  );

  const handleRemove = useCallback(
    (fileName: string) => {
      setJobMappings((prev) => {
        const newMappings = { ...prev };
        delete newMappings[fileName];
        return newMappings;
      });
    },
    [setJobMappings]
  );

  const handleError = useCallback((fileName: string, error: string) => {
    toast.error(`Unable to process input as job mappings: ${error}`, {
      description: fileName,
    });
  }, []);

  return (
    <div className="flex flex-col gap-2">
      <div>
        <SwitchCard
          isChecked={truncateAll}
          onCheckedChange={(value) => {
            setTruncateAll(value);
            if (overrideOverlapping) {
              setOverrideOverlapping(false);
            }
          }}
          title="Truncate existing mappings"
          description="Start fresh and only use what was imported."
        />
      </div>
      <div>
        <SwitchCard
          isChecked={overrideOverlapping}
          onCheckedChange={(value) => {
            setOverrideOverlapping(value);
            if (truncateAll) {
              setTruncateAll(false);
            }
          }}
          title="Override existing mappings"
          description="If yes, replaces existing mappings. If no, import new ones only."
        />
      </div>
      <FileUpload<JobMapping[]>
        validation={{
          validateType: validateFileType,
          validateSize: validateFileSize,
          maxSizeDisplay: MAX_FILE_SIZE_DISPLAY,
          acceptedTypes: '.json,application/json',
        }}
        onProcess={processFile}
        onSuccess={handleSuccess}
        onRemove={handleRemove}
        onError={handleError}
        renderFileExtra={(_, data) => (
          <span>{getFormattedCount(data?.length ?? 0)}</span>
        )}
      />
    </div>
  );
}

const US_NUMBER_FORMAT = new Intl.NumberFormat('en-US');
function getFormattedCount(count: number): string {
  return US_NUMBER_FORMAT.format(count);
}

function validateFileType(file: File): boolean {
  return file.type === 'application/json' || file.name.endsWith('.json');
}

const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50MB or 52.42MiB (52,427,800 bytes)
const MAX_FILE_SIZE_DISPLAY = filesize(MAX_FILE_SIZE, {
  standard: 'jedec', // Renders in MB, which users may be more familiar with
  round: 0,
});

// Handles byte comparison. File.size is the size of the file in bytes
function validateFileSize(file: File): boolean {
  return file.size <= MAX_FILE_SIZE;
}
