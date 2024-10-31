import { useState } from 'react';

interface FileDownloadProps<T> {
  data: T;
  fileName?: string;
  shouldFormat?: boolean;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

interface UseFileDownloadResponse<T> {
  downloadFile(props: FileDownloadProps<T>): Promise<void>;
  isDownloading: boolean;
  error?: string | null;
}

// Async Worker code that handles stringifying the input contents
const WORKER_CODE = `
  self.onmessage = function(e) {
    try {
      const { data, shouldFormat } = e.data;
      const jsonString = shouldFormat
        ? JSON.stringify(data, null, 2)
        : JSON.stringify(data);
      self.postMessage(jsonString);
    } catch (error) {
      self.postMessage({ error: error.message });
    }
  }
`;
const WORKER_BLOB = new Blob([WORKER_CODE], { type: 'application/javascript' });

/* Hook that provides ability to download a JSON file.
 */
export function useJsonFileDownload<T = unknown>(): UseFileDownloadResponse<T> {
  const [isDownloading, setIsDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function downloadFile({
    data,
    fileName = 'download.json',
    shouldFormat,
    onSuccess,
    onError,
  }: FileDownloadProps<T>): Promise<void> {
    setIsDownloading(true);
    setError(null);

    try {
      if (!data) {
        throw new Error('No data provided for download');
      }

      const workerUrl = URL.createObjectURL(WORKER_BLOB);
      const worker = new Worker(workerUrl);

      // Handle worker response with timeout
      const textToDownload = await Promise.race([
        new Promise<string>((resolve, reject) => {
          worker.onmessage = (e) => {
            if (e.data.error) {
              reject(new Error(e.data.error));
            } else {
              resolve(e.data);
            }
          };
          worker.onerror = (e) => reject(new Error(e.message));
          worker.postMessage({ data, shouldFormat });
        }),
        new Promise<never>((_, reject) =>
          setTimeout(
            () =>
              reject(
                new Error(
                  'Timed out while waiting to receive JSON stringification response'
                )
              ),
            10000
          )
        ),
      ]);

      // Clean up worker
      worker.terminate();
      URL.revokeObjectURL(workerUrl);

      // Create and download file
      await downloadContents(textToDownload, fileName);

      onSuccess?.();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Download failed';
      setError(errorMessage);
      onError?.(error instanceof Error ? error : new Error(errorMessage));
    } finally {
      setIsDownloading(false);
    }
  }

  return { downloadFile, isDownloading, error };
}

async function downloadContents(
  input: string,
  fileName: string
): Promise<void> {
  // Create download stream
  const stream = new ReadableStream({
    start(controller) {
      controller.enqueue(new TextEncoder().encode(input));
      controller.close();
    },
  });

  // Create and download file
  const response = new Response(stream);
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);

  // Trigger download
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;

  // Use click() directly instead of appending to document
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();

  // Cleanup
  setTimeout(() => {
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }, 100);
}
