import Image from 'next/image';
import { ReactElement } from 'react';

export default function JobSchedules(): ReactElement {
  return (
    <div className="px-6 pb-10 lg:pb-20">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Full Control Over Job Execution
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Configure jobs to run on a schedule or run them ad-hoc. Have full
        visibility into metrics and logs for each job.
      </div>
      <div className=" pt-10 lg:py-10 justify-center flex">
        <div className="border border-gray-400 rounded-xl overflow-hidden shadow-xl max-w-[1300px]">
          <Image
            src="https://assets.nucleuscloud.com/neosync/marketingsite/jobschedule.png"
            width="1300"
            height="700"
            alt="job"
          />
        </div>
      </div>
    </div>
  );
}
