import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import { ReactElement } from 'react';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import PrivateBetaForm from './PrivateBetaForm';

export default function PrivateBetaButton(): ReactElement {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="default" className="px-6">
          Neosync Cloud <ArrowRightIcon className="ml-2 h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg bg-white  p-6 shadow-xl">
        <DialogHeader>
          <div className="flex justify-center pt-10">
            <Image
              src="https://assets.nucleuscloud.com/neosync/newbrand/logo_text_light_mode.svg"
              alt="NeosyncLogo"
              width="118"
              height="30"
            />
          </div>
          <DialogTitle className="text-gray-900 text-2xl text-center pt-10">
            Join the Neosync Cloud Private Beta
          </DialogTitle>
          <DialogDescription className="pt-6 text-gray-900 text-md text-center">
            Want to use Neosync but don&apos;t want to host it yourself? Sign up
            for the Neosync Cloud Private Beta and get an environment.
          </DialogDescription>
        </DialogHeader>
        <div className="flex items-center space-x-2">
          <PrivateBetaForm />
        </div>
      </DialogContent>
    </Dialog>
  );
}
