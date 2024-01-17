import asyncio
import logging
from datetime import timedelta

from temporalio import workflow
from temporalio.client import Client
from temporalio.worker import Worker

with workflow.unsafe.imports_passed_through():
   from activities import (SampleModelInput, TrainModelInput, sample_model,
                           train_model)

@workflow.defn
class TrainWorkflow:
  @workflow.run
  async def run(self):
    workflow.logger.info("Running training workflow")
    return await workflow.execute_activity(
      train_model,
      TrainModelInput(
        10,
        ['age', 'workclass', 'education', 'education-num', 'marital-status', 'occupation', 'relationship', 'race', 'sex', 'native-country', 'income', 'hours-per-week'],
        '/Users/nick/code/nucleus/neosync/ml/backend/storage/adult.pkl',
        'postgresql://postgres:foofar@localhost:5434/nucleus?sslmode=disable',
        "public",
        "adult",
        ["*"],
      ),
      start_to_close_timeout=timedelta(seconds=10 * 60),
    )


@workflow.defn
class SampleWorkflow:
   @workflow.run
   async def run(self):
      workflow.logger.info("Running sample workflow")
      return await workflow.execute_activity(
        sample_model,
        SampleModelInput(
           10000,
           "/Users/nick/code/nucleus/neosync/ml/backend/storage/adult.pkl",
           "postgresql://postgres:foofar@localhost:5435/nucleus?sslmode=disable",
           "public",
           "adult",
        ),
        start_to_close_timeout=timedelta(seconds=10 * 60),
      )


async def main():
  client = await Client.connect("0.0.0.0:7233")

  worker = Worker(
      client,
      task_queue="synth-gen",
      workflows=[TrainWorkflow, SampleWorkflow],
      activities=[train_model, sample_model],
  )
  await worker.run()

if __name__ == "__main__":
    asyncio.run(main())
