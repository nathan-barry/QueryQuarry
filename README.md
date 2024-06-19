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
