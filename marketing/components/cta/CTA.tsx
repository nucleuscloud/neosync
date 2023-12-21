import { ReactElement } from 'react';
import WaitlistForm from '../buttons/WaitlistForm';

export default function CTA(): ReactElement {
  return (
    <div className="bg-gradient-to-tr from-[#0F0F0F] to-[#191919] my-20 border-2 border-gray-700 shadow-lg rounded-xl">
      <div className="flex flex-col align-center space-y-6 py-10 justify-center px-[25%]">
        <div className="text-gray-300 text-md font-satoshi text-xl text-center">
          Sign up to be notified when we launch
        </div>
        <WaitlistForm />
      </div>
    </div>
  );
}
