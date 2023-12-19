import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import BaseLayout from './BaseLayout';

export default function Component() {
  return (
    <BaseLayout>
      <div className="flex flex-col items-center justify-center h-screen bg-background">
        <Card className="w-full max-w-md p-8 dark:bg-background">
          <CardHeader>
            <CardTitle className="text-center text-gray-900 dark:text-gray-200 text-2xl">
              404 Error
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="text-center">
              <div className="text-gray-600 dark:text-gray-400">
                The page you are looking for does not exist or may have been
                moved.
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </BaseLayout>
  );
}
