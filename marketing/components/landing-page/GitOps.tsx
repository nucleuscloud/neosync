'use client';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';
import { SiTerraform } from 'react-icons/si';

export default function GitOpsSection(): ReactElement {
  const jobCode = `resource "neosync_job" "staging-sync-job" {
    name = "prod-to-stage"

    source_id = neosync_postgres_connection.prod_db.id
    destination_ids = [
      neosync_postgres_connection.stage_db.id,
      neosync_s3_connection.stage_backup.id,
    ]

    schedule = "0 23 * * *" # 11pm every night

    halt_on_new_column_addition = false

    mappings = [
      {
        "schema" : "public",
        "table" : "users",
        "column" : "account_number",
        "transformer" : "custom_account_number",
      },
      {
        "schema" : "public",
        "table" : "users",
        "column" : "address",
        "transformer" : "address_anonymize"
      },
    ]
  }`;

  return (
    <div className="px-6">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Synthetic Data meets GitOps
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Neosync is built with DevOps and infrastructure teams in mind. Use
        frameworks you know like terraform to manage your Neosync infrastructure
        and even create new jobs.
      </div>
      <div className="flex flex-col lg:flex-row items-center justify-center pt-20 gap-4">
        <div className="border border-gray-400 rounded-xl shadow-xl p-4">
          <Image
            src="/images/neosync-ci.svg"
            alt="pre"
            width="514"
            height="617"
            className="w-full"
          />

          <div className="pt-8">
            <GitHubLogoIcon className="text-gray-800 w-4 h-4" />
            <div className="font-sans font-bold text-gray-800 mb-2 mt-2">
              Neosync in CI
            </div>
            <div className="font-sans font-normal text-gray-800 text-sm">
              Use Neosync in your CI pipeline to hydrate CI databases with
              synthetic and anonymized data
            </div>
          </div>
        </div>
        <div className="border border-gray-400 rounded-xl shadow-xl p-4">
          <div>
            <Image
              src="/images/neosync-tf.svg"
              alt="pre"
              width="566"
              height="697"
            />
          </div>
          <div className="pt-8">
            <SiTerraform className="text-gray-800 w-4 h-4" />
            <div className="font-sans font-bold text-gray-800 mb-2 mt-2">
              Neosync Terraform Provider
            </div>
            <div className="font-sans font-normal text-gray-800 text-sm">
              Use the Neosync Terraform provider to manage your Neosync
              resources in code.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
