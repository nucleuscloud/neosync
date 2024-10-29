import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import SwitchCard from '@/components/switches/SwitchCard';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/libs/utils';
import { JobMapping } from '@neosync/sdk';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
  UploadIcon,
} from '@radix-ui/react-icons';
import {
  ChangeEvent,
  DragEvent,
  ReactElement,
  SetStateAction,
  useCallback,
  useEffect,
  useState,
} from 'react';
import { IoAlertCircleOutline, IoCloseCircle } from 'react-icons/io5';
import { toast } from 'sonner';

interface ImportConfig {
  truncateAll: boolean;
  overrideOverlapping: boolean;
}

interface Props {
  onImport(jobmappings: JobMapping[], config: ImportConfig): void;
}

function squashJobMappings(input: JobMapping[][]): JobMapping[] {
  return input.reduce((output: JobMapping[], curr: JobMapping[]) => {
    curr.forEach((mapping) => {
      // Check if we already have this combination
      const exists = output.some(
        (existing) =>
          existing.schema === mapping.schema &&
          existing.table === mapping.table &&
          existing.column === mapping.column
      );

      // Only add if it's not already present
      if (!exists) {
        output.push(mapping);
      }
    });

    return output;
  }, []);
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
  const [overrideOverallping, setOverrideOverlapping] = useState(false);
  return (
    <div>
      <ConfirmationDialog
        trigger={
          <Button type="button" variant="outline">
            <ButtonText text="Import" />
          </Button>
        }
        headerText="Import Job Mappings"
        description="This will import job mappings into the current job. Multiple files may be uploaded."
        body={
          <Body
            jobmappings={jmExtracted}
            setJobMappings={setJmExtracted}
            overrideOverlapping={overrideOverallping}
            setOverrideOverlapping={setOverrideOverlapping}
            truncateAll={truncateAll}
            setTruncateAll={setTruncateAll}
          />
        }
        containerClassName="max-w-xl"
        buttonText="Import"
        onConfirm={() => {
          const sortedEntries = Object.entries(jmExtracted).sort(
            ([, a], [, b]) => a.lastModified - b.lastModified
          );
          onImport(
            squashJobMappings(
              sortedEntries.map(([, values]) => values.mappings)
            ),
            {
              overrideOverlapping: overrideOverallping,
              truncateAll: truncateAll,
            }
          );
        }}
      />
    </div>
  );
}

// Web Worker code as a blob
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
  jobmappings: Record<string, ExtractedJobMappings>;
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
    jobmappings,
    setJobMappings,
    overrideOverlapping,
    setOverrideOverlapping,
    truncateAll,
    setTruncateAll,
  } = props;

  const [isDragging, setIsDragging] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
  const [errors, setErrors] = useState<Record<string, string | null>>({});
  const [processing, setProcessing] = useState<Record<string, boolean | null>>(
    {}
  );

  const [worker, setWorker] = useState<Worker | null>(null);

  useEffect(() => {
    const workerUrl = URL.createObjectURL(workerBlob);

    const newWorker = new Worker(workerUrl, { name: 'job-mappings-uploader' });

    newWorker.onmessage = (e) => {
      const { fileName, success, error, data } = e.data;

      if (!success) {
        setProcessing((prev) => ({ ...prev, [fileName]: false }));
        setErrors((prev) => ({
          ...prev,
          [fileName]: error,
        }));
      }
      try {
        if (Array.isArray(data)) {
          const jmappings = data.map((d) => JobMapping.fromJson(d));
          const file = files.find((f) => f.name === fileName);
          const newVal: ExtractedJobMappings = {
            mappings: jmappings,
            lastModified: file?.lastModified ?? 0,
          };
          setJobMappings((prev) => ({
            ...prev,
            [fileName]: newVal,
          }));
        }
      } catch (err) {
        toast.error(`Unable to process input as job mappings: ${err}`);
        setErrors((prev) => ({
          ...prev,
          [fileName]:
            err instanceof Error
              ? err.message
              : `unable to process mappings: ${err}`,
        }));
      } finally {
        setProcessing((prev) => ({ ...prev, [fileName]: false }));
      }
      console.log(`Processed ${fileName}: ${success ? 'success' : error}`);
    };

    setWorker(newWorker);

    return () => {
      newWorker.terminate();
      URL.revokeObjectURL(workerUrl);
    };
  }, []);

  const processFile = useCallback(
    (file: File) => {
      if (!validateFileType(file)) {
        setErrors((prev) => ({
          ...prev,
          [file.name]: 'Only JSON files are allowed',
        }));
        return;
      }

      if (!validateFileSize(file)) {
        setErrors((prev) => ({
          ...prev,
          [file.name]: 'File size must be less than 50MB',
        }));
        return;
      }

      setProcessing((prev) => ({ ...prev, [file.name]: true }));
      setErrors((prev) => ({ ...prev, [file.name]: null }));

      if (worker) {
        const processId = Date.now().toString() + Math.random().toString(36);
        worker.postMessage({ file, id: processId });
        console.log(`Started processing ${file.name} with ID ${processId}`);
      } else {
        console.error('Worker not initialized');
        setErrors((prev) => ({
          ...prev,
          [file.name]: 'Internal error: Worker not initialized',
        }));
        setProcessing((prev) => ({ ...prev, [file.name]: false }));
      }
    },
    [worker]
  );

  const onDragOver = useCallback((e: any) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const onDragLeave = useCallback((e: any) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const onDrop = useCallback(
    (e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      setIsDragging(false);
      const droppedFiles = Array.from(e.dataTransfer?.files ?? []);
      setFiles((prev) => [...prev, ...droppedFiles]);
      droppedFiles.forEach(processFile);
    },
    [processFile]
  );

  const onFileSelect = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const selectedFiles = Array.from(e.target.files ?? []);
      setFiles((prev) => [...prev, ...selectedFiles]);
      selectedFiles.forEach(processFile);
    },
    [processFile]
  );

  const removeFile = useCallback(
    (fileToRemove: File) => {
      setFiles(files.filter((file) => file !== fileToRemove));
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[fileToRemove.name];
        return newErrors;
      });
      setProcessing((prev) => {
        const newProcessing = { ...prev };
        delete newProcessing[fileToRemove.name];
        return newProcessing;
      });
      setJobMappings((prev) => {
        const newJmExtracted = { ...prev };
        delete newJmExtracted[fileToRemove.name];
        return newJmExtracted;
      });
    },
    [files]
  );
  return (
    <div className="flex flex-col gap-2">
      <div>
        <SwitchCard
          isChecked={truncateAll}
          onCheckedChange={setTruncateAll}
          title="Truncate existing mappings"
          description="This will clear any currently configured mappings and import everything that was found in the files."
        />
      </div>
      <div>
        <SwitchCard
          isChecked={overrideOverlapping}
          onCheckedChange={setOverrideOverlapping}
          title="Override any existing job mappings."
          description="If enabled, will override any existing mappings found in the table. Otherwise, the import will only add new mappings not found."
        />
      </div>
      <div
        className={cn(
          'border-2 border-dashed',
          isDragging ? 'border-blue-500' : 'border-gray-300'
        )}
      >
        <div
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
          className="flex flex-col items-center justify-center p-8 text-center cursor-pointer"
          onClick={() => document.getElementById('file-input')?.click()}
        >
          <UploadIcon className="w-12 h-12 mb-4 text-gray-400" />
          <h3 className="mb-2 text-lg font-semibold">Drop JSON Files Here</h3>
          <p className="mb-4 text-sm text-gray-500">or click to browse</p>
          <p className="text-xs text-gray-400">Maximum file size: 50MB</p>
          <Input
            id="file-input"
            type="file"
            multiple
            accept=".json,application/json"
            className="hidden"
            onChange={onFileSelect}
          />
        </div>
      </div>
      <div>
        {files.length > 0 && (
          <div className="mt-4 space-y-2">
            {files.map((file, index) => (
              <div key={index} className="flex flex-col space-y-2">
                <div className="flex items-center justify-between p-3 bg-white rounded-lg border">
                  <div className="flex items-center">
                    <div className="ml-2">
                      <p className="text-sm font-medium">{file.name}</p>
                      <p className="text-xs text-gray-500">
                        {(file.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {processing[file.name] && (
                      <div className="animate-spin rounded-full h-4 w-4 border-2 border-blue-500 border-t-transparent" />
                    )}
                    {!processing[file.name] && !errors[file.name] && (
                      <CheckCircledIcon className="w-4 h-4 text-green-500" />
                    )}
                    {!processing[file.name] && errors[file.name] && (
                      <ExclamationTriangleIcon className="w-4 h-4 text-red-500" />
                    )}
                    {!processing[file.name] &&
                      !errors[file.name] &&
                      jobmappings[file.name]?.mappings?.length > 0 && (
                        <span>{jobmappings[file.name].mappings.length}</span>
                      )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeFile(file)}
                      className="text-gray-500 hover:text-red-500"
                    >
                      <IoCloseCircle className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
                {errors[file.name] && (
                  <Alert variant="destructive">
                    <div className="flex flex-row items-center gap-2">
                      <IoAlertCircleOutline className="h-6 w-6" />
                      <AlertTitle className="font-semibold">
                        Issue with {file.name}
                      </AlertTitle>
                    </div>
                    <AlertDescription className="pl-8">
                      {errors[file.name]}
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function validateFileType(file: File): boolean {
  return file.type === 'application/json' || file.name.endsWith('.json');
}

function validateFileSize(file: File): boolean {
  const maxSize = 50 * 1024 * 1024; // 50MB
  return file.size <= maxSize;
}
