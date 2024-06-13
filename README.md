# query-quarry

This repo is for running a service that allows people to search whether a given text appears in a common LLM training dataset. Examples of potential uses are detecting copyrighted material, dataset contamination, etc.

## TODO
- [ ] Create the binary search function
    - [ ] Create data repo and figure out structure of files
    - [ ] Determine the length of the suffix array, size of pointers
    - [ ] Implement boolean contains function with off disk binary search
    - [ ] Refactor to count occurrences by splitting to do two binary search after detecting a copy
- [ ] Create a server that exposes search functionality
