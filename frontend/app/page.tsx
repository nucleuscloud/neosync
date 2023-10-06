'use client';
import RunsTable from './runs/components/RunsTable';

export default function Home() {
  return (
    <div className="space-y-4 my-2">
      <h1 className="text-lg font-semibold tracking-tight">Latest Job Runs</h1>
      <RunsTable />
    </div>
  );
}
