// FileUpload.tsx
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/libs/utils';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
  TrashIcon,
  UploadIcon,
} from '@radix-ui/react-icons';
import { format } from 'date-fns';
import { filesize } from 'filesize';
import {
  ChangeEvent,
  DragEvent,
  ReactElement,
  useCallback,
  useState,
} from 'react';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { LuFileJson } from 'react-icons/lu';

interface FileValidation {
  validateType(file: File): boolean;
  validateSize(file: File): boolean;
  maxSize?: number; // In bytes
  acceptedTypes?: string;
  maxSizeDisplay?: string;
}

export interface FileProcessingResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

interface FileUploadProps<T> {
  validation: FileValidation;
  onProcess: (file: File) => Promise<FileProcessingResult<T>>;
  onSuccess?: (fileName: File, data: T) => void;
  onError?: (fileName: string, error: string) => void;
  onRemove?: (fileName: string) => void;
  renderFileExtra?: (fileName: string, data?: T) => ReactElement<any>;
}

// Used for the hidden input
const fileUploadInputId = 'file-upload-input';

export function FileUpload<T>({
  validation,
  onProcess,
  onSuccess,
  onError,
  onRemove,
  renderFileExtra,
}: FileUploadProps<T>): ReactElement<any> {
  const [isDragging, setIsDragging] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [processing, setProcessing] = useState<Record<string, boolean>>({});
  const [processedData, setProcessedData] = useState<Record<string, T>>({});

  const processFile = useCallback(
    async (file: File) => {
      if (!validation.validateType(file)) {
        setErrors((prev) => ({
          ...prev,
          [file.name]: 'Invalid file type',
        }));
        return;
      }

      if (!validation.validateSize(file)) {
        setErrors((prev) => ({
          ...prev,
          [file.name]: `File size must be less than ${validation.maxSizeDisplay || 'maximum size'}`,
        }));
        return;
      }

      setProcessing((prev) => ({ ...prev, [file.name]: true }));
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[file.name];
        return newErrors;
      });

      try {
        const result = await onProcess(file);

        if (!result.success || !result.data) {
          const error = result.error || 'Processing failed';
          setErrors((prev) => ({ ...prev, [file.name]: error }));
          onError?.(file.name, error);
        } else {
          setProcessedData((prev) => ({ ...prev, [file.name]: result.data! }));
          onSuccess?.(file, result.data);
        }
      } catch (err) {
        const error = err instanceof Error ? err.message : 'Processing failed';
        setErrors((prev) => ({ ...prev, [file.name]: error }));
        onError?.(file.name, error);
      } finally {
        setProcessing((prev) => ({ ...prev, [file.name]: false }));
      }
    },
    [onProcess, onSuccess, onError, validation]
  );

  const onDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const onDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
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
      setProcessedData((prev) => {
        const newData = { ...prev };
        delete newData[fileToRemove.name];
        return newData;
      });
      onRemove?.(fileToRemove.name);
    },
    [files, onRemove]
  );

  return (
    <div className="flex flex-col gap-2">
      <div
        className={cn(
          'border-2 border-dashed rounded-lg bg-gray-100 dark:bg-transparent',
          isDragging
            ? 'border-blue-500 bg-blue-100 dark:bg-blue-700/10'
            : 'border-gray-300 dark:border-gray-500'
        )}
      >
        <div
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
          className="flex flex-col items-center justify-center p-8 text-center cursor-pointer"
          onClick={() => document.getElementById(fileUploadInputId)?.click()}
        >
          <UploadIcon className="w-12 h-12 mb-4 text-gray-400 dark:text-gray-600" />
          <h3 className="mb-2 text-lg font-semibold">Drop Files Here</h3>
          <p className="mb-4 text-sm text-gray-500">or click to browse</p>
          {validation.maxSizeDisplay && (
            <p className="text-xs text-gray-400">
              Maximum file size: {validation.maxSizeDisplay}
            </p>
          )}
          <Input
            id={fileUploadInputId}
            type="file"
            multiple
            accept={validation.acceptedTypes}
            className="hidden"
            onChange={onFileSelect}
          />
        </div>
      </div>
      <div>
        <UploadedFiles
          files={files}
          errors={errors}
          processing={processing}
          processedData={processedData}
          removeFile={removeFile}
          renderFileExtra={renderFileExtra}
        />
      </div>
    </div>
  );
}

interface UploadedFilesProps<T> {
  files: File[];
  processing: Record<string, boolean>;
  errors: Record<string, string>;
  processedData: Record<string, T>;
  renderFileExtra?: (fileName: string, data?: T) => ReactElement<any>;
  removeFile(fileToRemove: File): void;
}

function UploadedFiles<T>(props: UploadedFilesProps<T>): ReactElement<any> {
  const {
    files,
    processing,
    errors,
    processedData,
    renderFileExtra,
    removeFile,
  } = props;

  return (
    <>
      {files.length > 0 && (
        <div className="mt-4 space-y-2">
          {files.map((file, index) => (
            <div key={index} className="flex flex-col space-y-2">
              <div className="flex items-center justify-between p-3 dark:bg-gray-700 rounded-lg border dark:border-0">
                <div className="flex items-center">
                  <LuFileJson className="w-6 h-6" />
                  <div className="ml-2">
                    <p className="text-sm font-medium">{file.name}</p>
                    <div className="flex flex-row gap-2">
                      <p className="text-xs text-gray-500">
                        {filesize(file.size)}
                      </p>
                      <p className="text-xs text-gray-500">
                        {format(new Date(file.lastModified), 'PPpp')}
                      </p>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
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
                    renderFileExtra?.(file.name, processedData[file.name])}
                  <div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeFile(file)}
                      className="text-red-500  hover:text-red-500 px-1"
                    >
                      <TrashIcon className="w-4 h-4 mt-[1px]" />
                    </Button>
                  </div>
                </div>
              </div>
              {errors[file.name] && (
                <InvalidFileAlert
                  fileName={file.name}
                  error={errors[file.name]}
                />
              )}
            </div>
          ))}
        </div>
      )}
    </>
  );
}

interface InvalidFileAlertProps {
  fileName: string;
  error: string;
}
function InvalidFileAlert(props: InvalidFileAlertProps) {
  const { fileName, error } = props;
  return (
    <Alert variant="destructive">
      <div className="flex flex-row items-center gap-2">
        <IoAlertCircleOutline className="h-6 w-6" />
        <AlertTitle className="font-semibold">Issue with {fileName}</AlertTitle>
      </div>
      <AlertDescription className="pl-8">{error}</AlertDescription>
    </Alert>
  );
}
