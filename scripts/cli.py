import json
import os
import sys
import time
import base64
import argparse
import csv
from pathlib import Path

import requests

LOCALHOST = "http://localhost:8081/"

COUNT = "count"
CSV_ACTION = "csv"


def parse_arguments():
    parser = argparse.ArgumentParser(description="Query Client")
    parser.add_argument(
        "-action",
        type=str,
        default=COUNT,
        choices=[COUNT, CSV_ACTION],
        help="Choose action: 'count' or 'csv'",
    )
    parser.add_argument(
        "-data",
        type=str,
        default="./data/wiki40b.test",
        help="Path to dataset",
    )
    parser.add_argument(
        "-file",
        type=str,
        required=True,
        help="Path to file with queries",
    )
    parser.add_argument(
        "-tokenize",
        type=str,
        default="",
        help="Tokenizer name",
    )
    return parser.parse_args()


def create_request_payload(dataset, query, tokenize):
    return {
        "Dataset": dataset,
        "Length": len(query),
        "Query": query,
        "Tokenize": tokenize
    }


def cmd_count(client, queries, dataset, tokenizer_name):
    start_time = time.time()

    if tokenizer_name:
        from transformers import AutoTokenizer
        import numpy as np
        tokenizer = AutoTokenizer.from_pretrained(tokenizer_name)

    for query in queries:
        query = query.strip()
        if not query:
            continue
        print(f"{query}: ", end="")

        if tokenizer_name:
            byte_data = np.array(tokenizer.encode(query), dtype=np.uint16).view(np.uint8).tobytes()
            query = base64.b64encode(byte_data).decode('utf-8')

        payload = create_request_payload(dataset, query, tokenizer_name != "")
        try:
            response = client.post(
                f"{LOCALHOST}{COUNT}",
                json=payload,
                headers={"Content-Type": "application/json"},
            )
        except requests.RequestException as e:
            print()
            sys.exit(f"Error sending request: {e}")

        if response.status_code == 200:
            try:
                response_data = response.json()
                occurrences = response_data.get("occurrences", 0)
                print(occurrences)
            except json.JSONDecodeError:
                print()
                sys.exit("Error decoding JSON response")
        else:
            print()
            try:
                error_message = response.text
            except Exception:
                error_message = "No error message provided."
            sys.exit(
                f"Bad status code: {response.status_code}\nError Message: {error_message}"
            )

    end_time = time.time()
    print(f"Time Taken: {end_time - start_time:.2f} seconds")


def cmd_csv(client, queries, dataset, tokenizer_name, input_filename):
    start_time = time.time()

    input_path = Path(input_filename)
    output_filename = input_path.stem + "-results.csv"
    output_path = input_path.with_name(output_filename)

    # Increase the maximum field size limit
    csv.field_size_limit(sys.maxsize)

    try:
        with open(output_path, mode="w", newline="", encoding="utf-8") as out_file:
            writer = csv.writer(out_file)

            # Write CSV header
            writer.writerow(["queryID", "query", "docID", "document"])

            for i, query in enumerate(queries):
                query = query.strip()
                if not query:
                    continue
                print(f"{query}: ", end="")

                payload = create_request_payload(dataset, query, tokenizer_name != "")
                try:
                    response = client.post(
                        f"{LOCALHOST}{CSV_ACTION}",
                        json=payload,
                        headers={"Content-Type": "application/json"},
                    )
                except requests.RequestException as e:
                    sys.exit(f"Error sending request: {e}")

                if response.status_code == 200:
                    try:
                        csv_content = response.text
                        csv_reader = csv.reader(csv_content.splitlines())
                        records_written = 0
                        for record in csv_reader:
                            if len(record) < 2:
                                continue  # Skip invalid records
                            query_id = i
                            writer.writerow([query_id, query] + record)
                            records_written += 1
                        print("Successfully added to CSV")
                    except Exception as e:
                        sys.exit(f"Error processing CSV response: {e}")
                else:
                    try:
                        error_message = response.text
                    except Exception:
                        error_message = "No error message provided."
                    sys.exit(
                        f"\nBad status code: {response.status_code}\nError Message: {error_message}"
                    )

    except IOError as e:
        sys.exit(f"Error creating/writing to file: {e}")

    end_time = time.time()
    print(f"Time Taken: {end_time - start_time:.2f} seconds")


def main():
    args = parse_arguments()

    if args.action not in [COUNT, CSV_ACTION]:
        sys.exit("Invalid action. Choose 'count' or 'csv'.")

    dataset = args.data
    filename = args.file
    tokenizer = args.tokenize

    if not os.path.isfile(filename):
        sys.exit(f"Error opening the following file: {filename}")

    # Read all queries upfront to handle potential defer-like behavior
    try:
        with open(filename, mode="r", encoding="utf-8") as f:
            queries = f.readlines()
    except IOError as e:
        sys.exit(f"Error reading the file of queries: {e}")

    client = requests.Session()

    if args.action == COUNT:
        cmd_count(client, queries, dataset, tokenizer)
    elif args.action == CSV_ACTION:
        cmd_csv(client, queries, dataset, tokenizer, filename)


if __name__ == "__main__":
    main()
