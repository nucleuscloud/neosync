import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Link from 'next/link';
import { ReactElement } from 'react';
import PrivateBetaForm from '../buttons/PrivateBetaForm';
import { Button } from '../ui/button';

export default function CTA(): ReactElement {
  return (
    <div className="bg-[#F5F5F5] pb-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" bg-gradient-to-r from-slate-50 to-zinc-300 border-2 border-gray-700 shadow-lg rounded-xl">
          <div className="flex flex-col align-center space-y-6 py-10 justify-center px-[25%]">
            <div className="text-gray-900 text-4xl font-satoshi text-center">
              Get started with synthetic data at scale to build faster
            </div>
            <div className="flex flex-col lg:flex-row items-center justify-center gap-6">
              <div>
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="default">
                      Neosync Cloud <ArrowRightIcon className="ml-2 h-5 w-5" />
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-lg bg-black border border-gray-600 p-6">
                    <DialogHeader>
                      <DialogTitle className="text-white text-2xl">
                        Join the Neosync Cloud Private Beta
                      </DialogTitle>
                      <DialogDescription className="pt-10 text-gray-300 text-md">
                        Want to use Neosync but don&apos;t want to host it
                        yourself? Sign up for the private beta of Neosync Cloud.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="flex items-center space-x-2">
                      <PrivateBetaForm />
                    </div>
                    <DialogFooter className="sm:justify-start">
                      <DialogClose asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          className="text-white hover:bg-gray-800 hover:text-white"
                        >
                          Close
                        </Button>
                      </DialogClose>
                    </DialogFooter>
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
    </div>
  );
}
