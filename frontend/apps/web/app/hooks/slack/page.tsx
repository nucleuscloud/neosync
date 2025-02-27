'use client';
import { useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';

export default function SlackPage(): ReactElement {
  const searchParams = useSearchParams();

  const error = searchParams.get('error');
  const errorMessage = searchParams.get('error_message');

  if (error) {
    return (
      <div className="flex justify-center mt-24">
        <div className="w-full max-w-md p-6 bg-white rounded-lg shadow-md border border-red-200">
          <div className="flex flex-col items-center space-y-4">
            <div className="p-3 bg-red-100 rounded-full">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-8 w-8 text-red-500"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-800">
              Slack Connection Failed
            </h2>
            <div className="text-center">
              <p className="text-gray-600 mb-2">Error: {error}</p>
              {errorMessage && (
                <p className="text-gray-600">Error Message: {errorMessage}</p>
              )}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex justify-center mt-24">
      <div className="w-full max-w-md p-6 bg-white rounded-lg shadow-md border border-green-200">
        <div className="flex flex-col items-center space-y-4">
          <div className="p-3 bg-green-100 rounded-full">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-8 w-8 text-green-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-gray-800">
            Slack Connected Successfully
          </h2>
          <p className="text-gray-600 text-center">
            Your Slack workspace has been successfully connected. You can now
            close this window and return to the application.
          </p>
        </div>
      </div>
    </div>
  );
}
