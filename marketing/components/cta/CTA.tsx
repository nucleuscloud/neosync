import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import PrivateBetaForm from '../buttons/PrivateBetaForm';
import { Button } from '../ui/button';

export default function CTA(): ReactElement {
  return (
    <div>
      <div className=" bg-gradient-to-r from-slate-50 to-zinc-300 border border-gray-400 shadow-xl rounded-xl">
        <div className="flex flex-col align-center space-y-6 py-10 justify-center  px-[5%] lg:px-[25%]">
          <div className="text-gray-900 text-4xl font-satoshi text-center">
            Get started with synthetic data at scale
          </div>
          <div className="flex flex-col lg:flex-row items-center justify-center gap-6">
            <div>
              <Dialog>
                <DialogTrigger asChild>
                  <Button variant="default">
                    Neosync Cloud <ArrowRightIcon className="ml-2 h-5 w-8" />
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-lg bg-white p-6 shadow-xl">
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
                      Want to use Neosync but don&apos;t want to host it
                      yourself? Sign up for the private beta of Neosync Cloud
                      and get an environment.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="flex items-center space-x-2">
                    <PrivateBetaForm />
                  </div>
                </DialogContent>
              </Dialog>
            </div>
            <Button className="px-6" variant="secondary">
              <Link href="https://github.com/nucleuscloud/neosync">
                <div className="flex flex-row gap-2">
                  <GitHubLogoIcon className="mr-2 h-5 w-5" /> Open Source
                </div>
              </Link>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
