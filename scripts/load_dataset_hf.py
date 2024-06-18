# Copyright 2021 Google LLC
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# EDIT: Nathan Barry
# Removed the tokenize preprocessing option from the original script.
# Less useful for exact string match as it makes querying more complex.

import datasets
import os
import struct
import numpy as np
from tqdm import tqdm
import glob

import argparse


FILE_EXTENSIONS = {"text": "txt", "json": "jsonl", "csv": "csv"}

parser = argparse.ArgumentParser(description='Load a dataset.')
parser.add_argument('--save_dir', type=str, default="data/")
parser.add_argument('--name', type=str)
parser.add_argument('--data_dir', type=str, default=None)
parser.add_argument('--split', type=str)
parser.add_argument('--subset', type=str, default=None)
parser.add_argument('--num_workers', type=int, default=None)
parser.add_argument('--text_feature_key', type=str, default="text")

args = parser.parse_args()

save_dir = args.save_dir
data_dir = args.data_dir
dataset_name = args.name
split = args.split
subset = args.subset
num_workers = args.num_workers
key = args.text_feature_key

if dataset_name in FILE_EXTENSIONS:
    assert data_dir is not None
    data_files = glob.glob(f"{data_dir}/*.{FILE_EXTENSIONS[dataset_name]}")
    ds = datasets.load_dataset(dataset_name, subset, data_files=data_files, split=split)
else:
    ds = datasets.load_dataset(dataset_name, subset, split=split)
assert isinstance(ds, datasets.Dataset), "This is not a HF-dataset. It might be a DatasetDict. Try passing `split`?"

UID = 0


def sep():
    global UID
    UID += 1
    return b"\xff\xff" + struct.pack("<I", UID)

os.makedirs(save_dir, exist_ok=True)
fout = open(os.path.join(save_dir, dataset_name + "." + split), "wb")
sizes = [0]

for example in tqdm(ds):
    out = example[key].encode("utf8")
    next_line = sep() + out
    fout.write(next_line)
    sizes.append(sizes[-1] + len(next_line))

open(os.path.join(save_dir, dataset_name + "." + split + ".size"), "wb").write(
    np.array(sizes, dtype=np.uint64).tobytes())
