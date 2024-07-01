# QueryQuarry

This repo is for running a service that allows people to search whether a given text appears in a common LLM training dataset. Potential uses are dataset contamination research, detecting copyrighted material, personal identification, etc.



## Quickstart

### Constructing the Suffix Array
To download and preprocess a HuggingFace dataset (in this case wiki40b):
```
python3 scripts/load_dataset.py --name wiki40b --subset en --split test
```
This creates two files: `data/wiki40b.test` and `data/wiki40b.test.size`. The former is a preprocessed version of the dataset. It's essentially just all documents concatenated into one massive string, with document IDs placed in between each document. The latter is a file of sequential int64s with each representing a document is (ex. the 18th int64 is the length of document 18 in the text file). This file is important for quickly extracting the document.

To construct the suffix array:
```
python3 scripts/make_suffix_array.py data/wiki40b.test
```
This creates the file `data/wiki40b.test.table.bin`. This is the actual suffix array. The rust implementation automatically uses the minimum amount of bytes needed to index the text file. This is used for m*log(n) time search on the text, where n is the size of the text and m is the length of the query.
For ultra large files, you might need to run it with `ulimit -Sn 100000` to increase the number of files than can be simultaneously open on your machine (as construction opens many), otherwise the program might crash.


### Running the Server
To start the server:
```
go run cmd/server/main.go
```
You can see an interactive front-end by going to `http://localhost:8080`.


### Using the Client CLI
To count all occurrences for all queries in a file on a dataset:
```
go run cmd/cli/main.go --action count --data data/wiki40b.test --file queries/presidents.txt
```
As an example, we have a file containing the name of every U.S. President in the queries folder.

To generate a CSV with all the documents where an occurrence occurred on:
```
go run cmd/cli/main.go --action csv --data data/wiki40b.test --file queries/presidents.txt
```
This command creates a new CSV file named `presidents-results.csv`

To view the CSV in a jupyter notebok for this :
```
jupyter notebook queries/presidents-results.ipynb
```



## Why This Project?
Data contamination is a large issue in the current LLM space. In the paper <a href="https://aclanthology.org/2023.trustnlp-1.5/">Can we trust the evaluation on ChatGPT?</a>, the authors beg the question:

> "Given that ChatGPT is a closed model without [publicly available] information about its training dataset and how it is currently being trained, there is a large loxodonta mammal in the room: how can we know whether ChatGPT has not been contaminated with the evaluation datasets?"

This question affects every NLP researcher and was a problem I ran into time and time again when doing my own evaluations.

The papers [What's in My Big Data?](https://arxiv.org/abs/2310.20707) (AI2) and [Scalable Extraction of Training Data from (Production) Language Models](https://arxiv.org/abs/2311.17035) (Google Research) both constructed data structures on large corpa that allowed for efficient exact string lookup. Both papers required Google Cloud compute nodes with (882GB RAM, 224 CPUs) and (1.4TB RAM, 176 CPUs) respectively, expensive hardware that is out of reach for most researchers.

While the AI2 paper allows access to their search functionality, it is restricted due to the expensive nature of ElasticSearch and the Inverted Index data structure they used which, while more flexible, requires high memory and many cores to do search efficiently. The Google Research paper used Suffix Arrays which perform the lookup using binary search, which is a sequential algorithm and thus does not require much RAM to perform.

The Google Research paper used code from a previous Google Research paper, [Deduplicating Training Data Makes Language Models Better](https://arxiv.org/abs/2107.06499), to construct the Suffix Array. While this implementation does construct it using external memory, it does have the requirement that the initial dataset can fit into memory, thus still requiring expensive machines to construct the data structures.</p>

This project aims to host search functionality on a variety of popular large datasets. We use the aforementioned [Google Research repository](https://github.com/google-research/deduplicate-text-datasets) to construct these large data structures on any HuggingFace dataset. This project implements a more efficient Count Occurrences implementation that runs roughly 500-1000x faster than the original implementation. We provide a workflow that allows one to easily download a dataset, pre-process it, construct a suffix array, and host a service that allows one to query it. We provide a front-end for easy interaction and a CLI that can retrieve every document in a dataset where a query occurs.
