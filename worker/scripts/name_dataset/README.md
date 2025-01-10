# Names Dataset

These python scripts are used to generate the `first_names.txt` and `last_names.txt` that ultimately
end up in the worker datasets internal directory and must be manually copied there after generation.

This is a pretty hacky setup for now and is here for posterity or if we want to re-run wtih more granularity in the future.

## Env Setup

```console
python3 -m venv env
source bin/env/activate
pip3 install -r requirements.txt
```

## Names Dataset

### Run

There are different python scripts to run.

The first, `generate_names.py` pulls names out of the `NameDataset` and stores them in pickle files.
This can take a while depending on how much data you want which is why it is split into its own script.

The second script, `generate_text.py` loads these pickle files and invokes further processing on the data
before ultimately writing each name flatly into their respective text files.
This is useful if you want to further sort, filter, etc the original dataset.

Today we mostly just filter out non-ascii names and make sure each value is unique.

```console
python3 generate_names.py
python3 generate_text.py
```

### Copy

The output files should be copied to their respective locations where `go generate` can pick them up.

```console
cp first_names.txt ../internal/benthos/transformers/data-sets/datasets/first_names.txt
cp last_names.txt ../internal/benthos/transformers/data-sets/datasets/last_names.txt
```
