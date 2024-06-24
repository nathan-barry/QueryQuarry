# QueryQuarry

This repo is for running a service that allows people to search whether a given text appears in a common LLM training dataset. Examples of potential uses are detecting copyrighted material, dataset contamination, etc.

## TODO
- [X] Implement disk binary search
- [X] Implement server and API
- [X] Implement the front end 
- [X] Implement CLI tool
- [X] Implement CSV endpoint
- [ ] Add documentation to README
- [ ] Add graceful error handling
- [ ] Add markdown file with available datasets and crowdsourced insights for each 

## Quickstart

To download and preprocess a huggingface dataset:
```
python3 scripts/load_dataset.py --name wiki40b --subset en --split test
```

To construct the suffix array:
```
python3 scripts/make_suffix_array.py data/wiki40b.test
```

To start the server:
```
go run cmd/server/main.go
```

You can see an interactive front-end by going to `http://localhost:8080`.

In the queries folder, we have a file containing the name of every U.S. President.

To query using the client CLI:
```
go run cmd/cli/main.go -file ./queries/presidents.txt
```

This prints the number of times each President's name has occurred in the Wiki40b test set.

To generate a CSV with all the documents where an occurrence occurred:

```
go run cmd/cli/main.go -action csv -file ./queries/presidents.txt
```
