import Image from 'next/image';
import { ReactElement } from 'react';

export default function TransformerSection(): ReactElement {
  return (
    <div className="px-6 pb-10 lg:pb-20">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Full Control Over Data Transformation
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Configure Transformers to anonymize data or generate synthetic data.
        Neosync handles referential integrity automatically.
      </div>
      <div className=" pt-10 lg:py-10">
        <div className=" justify-center flex">
          <Image
            src="https://assets.nucleuscloud.com/neosync/marketingsite/schemapage.png"
            width="1300"
            height="700"
            alt="job"
            className="border border-gray-400 rounded-xl overflow-hidden shadow-xl"
          />
        </div>
      </div>
    </div>
  );
}
