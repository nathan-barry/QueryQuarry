# QueryQuarry

This repo is for running a service that allows people to search whether a given text appears in a common LLM training dataset. Examples of potential uses are detecting copyrighted material, dataset contamination, etc.

## TODO
- [X] Implement disk binary search
- [X] Implement server and API
- [X] Implement the front end 
- [X] Implement CLI tool
- [X] Add more features to API
    - [X] Grab the surrounding sentences of each occurrence
    - [X] Find all document IDs and positions of each document with an occurrence 
        - [ ] Send as CSV file
- [ ] Add documentation to README
- [ ] Add graceful error handling
- [ ] Add markdown file with available models and crowdsourced insights for each 


## Commands

To download and preprocess a huggingface dataset:
```
python3 scripts/load_dataset.py --name tiny_shakespeare --split test
python3 scripts/load_dataset.py --name wiki40b --subset en  --split test
```

To construct the suffix array:
```
python3 scripts/make_suffix_array.py data/tiny_shakespeare.test
python3 scripts/make_suffix_array.py data/wiki40b.test
```

To start the server:
```
go run cmd/server/main.go
```

To query using the client CLI:
```
go run cmd/cli/main.go --file="./queries/YOUR_FILE"
```
