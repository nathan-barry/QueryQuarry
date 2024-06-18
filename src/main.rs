/* Copyright 2021 Google LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* EDIT: Nathan Barry
 * This file is from the google-research deduplicate-text-dataset repo.
 * I've removed all functionality that doesn't directly relate to suffix
 * array construction and minimized dependencies. Originally I just linked
 * to the other repo, but thought it would be more user friendly if I just
 * included it in my repo and had scripts that automatically put everything
 * in the right place.
 */

/* Create and use suffix arrays for deduplicating language model datasets.
 *
 * A suffix array A for a sequence S is a datastructure that contains all
 * suffixes of S in sorted order. To be space efficient, instead of storing
 * the actual suffix, we just store the pointer to the start of the suffix.
 * To be time efficient, it uses fancy algorithms to not require quadratic
 * (or worse) work. If we didn't care about either, then we could literally
 * just define (in python)
 * A = sorted(S[i:] for i in range(len(S)))
 *
 * Suffix arrays are amazing because they allow us to run lots of string
 * queries really quickly, while also only requiring an extra 8N bytes of
 * storage (one 64-bit pointer for each byte in the sequence).
 *
 * This code is designed to work with Big Data (TM) and most of the
 * complexity revolves around the fact that we do not require the
 * entire suffix array to fit in memory. In order to keep things managable,
 * we *do* require that the original string fits in memory. However, even
 * the largest language model datasets (e.g., C4) are a few hundred GB
 * which on todays machines does fit in memory.
 *
 * With all that amazing stuff out of the way, just a word of warning: this
 * is the first program I've ever written in rust. I still don't actually
 * understand what borrowing something means, but have found that if I
 * add enough &(&&x.copy()).clone() then usually the compiler just loses
 * all hope in humanity and lets me do what I want. I apologize in advance
 * to anyone actually does know rust and wants to lock me in a small room
 * with the Rust Book by Klabnik & Nichols until I repent for my sins.
 * (It's me, two months in the future. I now more or less understand how
 * to borrow. So now instead of the code just being all awful, you'll get
 * a nice mix of sane rust and then suddenly OH-NO-WHAT-HAVE-YOU-DONE-WHY!?!)
 */

use std::fs;
use std::fs::File;
use std::io::prelude::*;
use std::io::BufReader;
use std::io::Read;
use std::time::Instant;

extern crate clap;
extern crate crossbeam;

use clap::{Parser, Subcommand};
use std::cmp::Ordering;
use std::collections::BinaryHeap;

mod table;

#[derive(Parser, Debug)]
#[clap(author, version, about, long_about = None)]
struct Args {
    #[clap(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[clap(arg_required_else_help = true)]
    Make {
        #[clap(short, long)]
        data_file: String,
    },

    MakePart {
        #[clap(short, long)]
        data_file: String,
        #[clap(short, long)]
        start_byte: usize,
        #[clap(short, long)]
        end_byte: usize,
    },

    Merge {
        #[clap(short, long)]
        suffix_path: Vec<String>,
        #[clap(short, long)]
        output_file: String,
        #[clap(short, long, default_value_t = 8)]
        num_threads: i64,
    },
}

/* Convert a uint64 array to a uint8 array.
 * This doubles the memory requirements of the program, but in practice
 * we only call this on datastructures that are smaller than our assumed
 * machine memory so it works.
 */
pub fn to_bytes(input: &[u64], size_width: usize) -> Vec<u8> {
    let mut bytes = Vec::with_capacity(size_width * input.len());

    for value in input {
        bytes.extend(&value.to_le_bytes()[..size_width]);
    }
    bytes
}

/* Convert a uint8 array to a uint64. Only called on (relatively) small files. */
pub fn from_bytes(input: Vec<u8>, size_width: usize) -> Vec<u64> {
    println!("S {}", input.len());
    assert!(input.len() % size_width == 0);
    let mut bytes: Vec<u64> = Vec::with_capacity(input.len() / size_width);

    let mut tmp = [0u8; 8];
    // todo learn rust macros, hope they're half as good as lisp marcos
    // and if they are then come back and optimize this
    for i in 0..input.len() / size_width {
        tmp[..size_width].copy_from_slice(&input[i * size_width..i * size_width + size_width]);
        bytes.push(u64::from_le_bytes(tmp));
    }

    bytes
}

/* For a suffix array, just compute A[i], but load off disk because A is biiiiiiigggggg. */
fn table_load_disk(table: &mut BufReader<File>, index: usize, size_width: usize) -> usize {
    table
        .seek(std::io::SeekFrom::Start((index * size_width) as u64))
        .expect("Seek failed!");
    let mut tmp = [0u8; 8];
    table.read_exact(&mut tmp[..size_width]).unwrap();
    return u64::from_le_bytes(tmp) as usize;
}

/* Binary search to find where query happens to exist in text */
fn off_disk_position(
    text: &[u8],
    table: &mut BufReader<File>,
    query: &[u8],
    size_width: usize,
) -> usize {
    let (mut left, mut right) = (0, text.len());
    while left < right {
        let mid = (left + right) / 2;
        if query < &text[table_load_disk(table, mid, size_width)..] {
            right = mid;
        } else {
            left = mid + 1;
        }
    }
    left
}

/*
 * We're going to work with suffix arrays that are on disk, and we often want
 * to stream them top-to-bottom. This is a datastructure that helps us do that:
 * we read 1MB chunks of data at a time into the cache, and then fetch new data
 * when we reach the end.
 */
struct TableStream {
    file: BufReader<File>,
    cache: [u8; 8],
    size_width: usize,
}

/* Make a table from a file path and a given offset into the table */
fn make_table(path: std::string::String, offset: usize, size_width: usize) -> TableStream {
    let mut table = TableStream {
        file: std::io::BufReader::with_capacity(1024 * 1024, fs::File::open(path).unwrap()),
        cache: [0u8; 8],
        size_width: size_width,
    };
    table
        .file
        .seek(std::io::SeekFrom::Start((offset * size_width) as u64))
        .expect("Seek failed!");
    return table;
}

/* Get the next word from the suffix table. */
fn get_next_pointer_from_table_canfail(tablestream: &mut TableStream) -> u64 {
    let ok = tablestream
        .file
        .read_exact(&mut tablestream.cache[..tablestream.size_width]);
    let bad = match ok {
        Ok(_) => false,
        Err(_) => true,
    };
    if bad {
        return std::u64::MAX;
    }
    let out = u64::from_le_bytes(tablestream.cache);
    return out;
}

/*
 * Create a suffix array for a given file in one go.
 * Calling this method is memory heavy---it's technically linear in the
 * length of the file, but the constant is quite big.
 * As a result, this method should only be called for files that comfortably
 * fit into memory.
 *
 * The result of calling this method is a new file with ".table.bin" appended
 * to the name which is the suffix array of sorted suffix pointers. This file
 * should be at most 8x larger than the original file (one u64 pointer per
 * byte of the original). In order to save space, if it turns out we only need
 * 32 bits to uniquely point into the data file then we serialize using fewer
 * bits (or 24, or 40, or ...), but in memory we always use a u64.
 *
 * If the file does not fit into memory, then instead you should use the
 * alternate save_part and then merge_parallel in two steps. See the comments
 * below for how those work.
 */
fn cmd_make(fpath: &String) -> std::io::Result<()> {
    let now = Instant::now();
    println!(
        "Reading the dataset at time t={}ms",
        now.elapsed().as_millis()
    );
    let mut text_ = Vec::with_capacity(std::fs::metadata(fpath.clone()).unwrap().len() as usize);
    fs::File::open(fpath.clone())
        .unwrap()
        .read_to_end(&mut text_)?;
    let text = &text_;
    println!(
        "Done reading the dataset at time t={}ms",
        now.elapsed().as_millis()
    );

    println!("... and now starting the suffix array construction.");

    let st = table::SuffixTable::new(text);
    println!(
        "Done building suffix array at t={}ms",
        now.elapsed().as_millis()
    );
    let parts = st.into_parts();
    let table = parts.1;

    let ratio = ((text.len() as f64).log2() / 8.0).ceil() as usize;
    println!("Ratio: {}", ratio);

    let mut buffer = File::create(fpath.clone() + ".table.bin")?;
    let bufout = to_bytes(&table, ratio);
    println!(
        "Writing the suffix array at time t={}ms",
        now.elapsed().as_millis()
    );
    buffer.write_all(&bufout)?;
    println!("And finished at time t={}ms", now.elapsed().as_millis());
    Ok(())
}

/*
 * Create a suffix array for a subsequence of bytes.
 * As with save, this method is linear in the number of bytes that are
 * being saved but the constant is rather high. This method does exactly
 * the same thing as save except on a range of bytes.
 */
fn cmd_make_part(fpath: &String, start: u64, end: u64) -> std::io::Result<()> {
    let now = Instant::now();
    println!("Opening up the dataset files");

    let space_available = std::fs::metadata(fpath.clone()).unwrap().len() as u64;
    assert!(start < end);
    assert!(end <= space_available);

    let mut text_ = vec![0u8; (end - start) as usize];
    let mut file = fs::File::open(fpath.clone()).unwrap();
    println!("Loading part of file from byte {} to {}", start, end);
    file.seek(std::io::SeekFrom::Start(start))
        .expect("Seek failed!");
    file.read_exact(&mut text_).unwrap();
    let text = &text_;
    println!(
        "Done reading the dataset at time t={}ms",
        now.elapsed().as_millis()
    );
    println!("... and now starting the suffix array construction.");

    let st = table::SuffixTable::new(text);
    println!(
        "Done building suffix array at t={}ms",
        now.elapsed().as_millis()
    );
    let parts = st.into_parts();
    let table = parts.1;

    let ratio = ((text.len() as f64).log2() / 8.0).ceil() as usize;
    println!("Ratio: {}", ratio);

    let mut buffer = File::create(format!("{}.part.{}-{}.table.bin", fpath, start, end))?;
    let mut buffer2 = File::create(format!("{}.part.{}-{}", fpath, start, end))?;
    let bufout = to_bytes(&table, ratio);
    println!(
        "Writing the suffix array at time t={}ms",
        now.elapsed().as_millis()
    );
    buffer.write_all(&bufout)?;
    buffer2.write_all(text)?;
    println!("And finished at time t={}ms", now.elapsed().as_millis());
    Ok(())
}

/*
 * A little bit of state for the merge operation below.
 * - suffix is suffix of one of the parts of the dataset we're merging;
this is the value we're sorting on
 * - position is the location of this suffix (so suffix = array[position..])
 * - table_index says which suffix array this suffix is a part of
 */
#[derive(Copy, Clone, Eq, PartialEq)]
struct MergeState<'a> {
    suffix: &'a [u8],
    position: u64,
    table_index: usize,
}

impl<'a> Ord for MergeState<'a> {
    fn cmp(&self, other: &Self) -> Ordering {
        other.suffix.cmp(&self.suffix)
    }
}

impl<'a> PartialOrd for MergeState<'a> {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

/*
 * Merge together M different suffix arrays (probably created with make-part).
 * That is, given strings S_i and suffix arrays A_i compute the suffix array
 * A* = make-suffix-array(concat S_i)
 * In order to do this we just implement mergesort's Merge operation on each
 * of the arrays A_i to construct a sorted array A*.
 *
 * This algorithm is *NOT A LINEAR TIME ALGORITHM* in the worst case. If you run
 * it on a dataset consisting entirely of the character A it will be quadratic.
 * Fortunately for us, language model datasets typically don't just repeat the same
 * character a hundred million times in a row. So in practice, it's linear time.
 *
 * There are thre complications here.
 *
 * As with selfsimilar_parallel, we can't fit all A_i into memory at once, and
 * we want to make things fast and so parallelize our execution. So we do the
 * same tricks as before to make things work.
 *
 * However we have one more problem. In order to know how to merge the final
 * few bytes of array S_0 into their correct, we need to know what bytes come next.
 * So in practice we make sure that S_{i}[-HACKSIZE:] === S_{i+1}[:HACKSIZE].
 * As long as HACKSIZE is longer than the longest potential match, everything
 * will work out correctly. (I did call it hacksize after all.....)
 * In practice this works. It may not for your use case if there are long duplicates.
 */
fn cmd_merge(
    data_files: &Vec<String>,
    output_file: &String,
    num_threads: i64,
) -> std::io::Result<()> {
    // This value is declared here, but also in scripts/make_suffix_array.py
    // If you want to change it, it needs to be changed in both places.
    const HACKSIZE: usize = 100000;

    let nn: usize = data_files.len();

    fn load_text2<'s, 't>(fpath: String) -> Vec<u8> {
        println!("Setup buffer");
        let mut text_ =
            Vec::with_capacity(std::fs::metadata(fpath.clone()).unwrap().len() as usize);
        println!("Done buffer {}", text_.len());
        fs::File::open(fpath.clone())
            .unwrap()
            .read_to_end(&mut text_)
            .unwrap();
        println!("Done read buffer");
        return text_;
    }

    // Start out by loading the data files and suffix arrays.
    let texts: Vec<Vec<u8>> = (0..nn).map(|x| load_text2(data_files[x].clone())).collect();

    let texts_len: Vec<usize> = texts
        .iter()
        .enumerate()
        .map(|(i, x)| x.len() - (if i + 1 == texts.len() { 0 } else { HACKSIZE }))
        .collect();

    let metadatas: Vec<u64> = (0..nn)
        .map(|x| {
            let meta = fs::metadata(format!("{}.table.bin", data_files[x].clone())).unwrap();
            assert!(meta.len() % (texts[x].len() as u64) == 0);
            return meta.len();
        })
        .collect();

    let big_ratio = ((texts_len.iter().sum::<usize>() as f64).log2() / 8.0).ceil() as usize;
    println!("Ratio: {}", big_ratio);

    let ratio = metadatas[0] / (texts[0].len() as u64);

    fn worker(
        texts: &Vec<Vec<u8>>,
        starts: Vec<usize>,
        ends: Vec<usize>,
        texts_len: Vec<usize>,
        part: usize,
        output_file: String,
        data_files: Vec<String>,
        ratio: usize,
        big_ratio: usize,
    ) {
        let nn = texts.len();
        let mut tables: Vec<TableStream> = (0..nn)
            .map(|x| make_table(format!("{}.table.bin", data_files[x]), starts[x], ratio))
            .collect();

        let mut idxs: Vec<u64> = starts.iter().map(|&x| x as u64).collect();

        let delta: Vec<u64> = (0..nn)
            .map(|x| {
                let pref: Vec<u64> = texts[..x].iter().map(|y| y.len() as u64).collect();
                pref.iter().sum::<u64>() - (HACKSIZE * x) as u64
            })
            .collect();

        let mut next_table = std::io::BufWriter::new(
            File::create(format!("{}.table.bin.{:04}", output_file.clone(), part)).unwrap(),
        );

        fn get_next_maybe_skip(
            mut tablestream: &mut TableStream,
            index: &mut u64,
            thresh: usize,
        ) -> u64 {
            //println!("{}", *index);
            let mut location = get_next_pointer_from_table_canfail(&mut tablestream);
            if location == u64::MAX {
                return location;
            }
            *index += 1;
            while location >= thresh as u64 {
                location = get_next_pointer_from_table_canfail(&mut tablestream);
                if location == u64::MAX {
                    return location;
                }
                *index += 1;
            }
            return location;
        }

        let mut heap = BinaryHeap::new();

        for x in 0..nn {
            let position = get_next_maybe_skip(&mut tables[x], &mut idxs[x], texts_len[x]);
            //println!("{} @ {}", position, x);
            heap.push(MergeState {
                suffix: &texts[x][position as usize..],
                position: position,
                table_index: x,
            });
        }

        // Our algorithm is not linear time if there are really long duplicates
        // found in the merge process. If this happens we'll warn once.
        let mut did_warn_long_sequences = false;

        let mut prev = &texts[0][0..];
        while let Some(MergeState {
            suffix: _suffix,
            position,
            table_index,
        }) = heap.pop()
        {
            //next_table.write_all(&(position + delta[table_index] as u64).to_le_bytes()).expect("Write OK");
            next_table
                .write_all(&(position + delta[table_index] as u64).to_le_bytes()[..big_ratio])
                .expect("Write OK");

            let position = get_next_maybe_skip(
                &mut tables[table_index],
                &mut idxs[table_index],
                texts_len[table_index],
            );
            if position == u64::MAX {
                continue;
            }

            if idxs[table_index] <= ends[table_index] as u64 {
                let next = &texts[table_index][position as usize..];
                //println!("  {:?}", &next[..std::cmp::min(10, next.len())]);

                let match_len = (0..50000000)
                    .find(|&j| !(j < next.len() && j < prev.len() && next[j] == prev[j]));
                if !did_warn_long_sequences {
                    if let Some(match_len_) = match_len {
                        if match_len_ > 5000000 {
                            println!("There is a match longer than 50,000,000 bytes.");
                            println!("You probably don't want to be using this code on this dataset---it's (possibly) quadratic runtime now.");
                            did_warn_long_sequences = true;
                        }
                    } else {
                        println!("There is a match longer than 50,000,000 bytes.");
                        println!("You probably don't want to be using this code on this dataset---it's quadratic runtime now.");
                        did_warn_long_sequences = true;
                    }
                }

                heap.push(MergeState {
                    suffix: &texts[table_index][position as usize..],
                    position: position,
                    table_index: table_index,
                });
                prev = next;
            }
        }
    }

    // Make sure we have enough space to take strided offsets for multiple threads
    // This should be an over-approximation, and starts allowing new threads at 1k of data
    //let num_threads = std::cmp::min(num_threads, std::cmp::max((texts.len() as i64 - 1024)/10, 1));
    println!("AA {}", num_threads);

    // Start a bunch of jobs that each work on non-overlapping regions of the final resulting suffix array
    // Each job is going to look at all of the partial suffix arrays to take the relavent slice.
    let _answer = crossbeam::scope(|scope| {
        let mut tables: Vec<BufReader<File>> = (0..nn)
            .map(|x| {
                std::io::BufReader::new(
                    fs::File::open(format!("{}.table.bin", data_files[x])).unwrap(),
                )
            })
            .collect();

        let mut starts = vec![0; nn];

        for i in 0..num_threads as usize {
            let texts = &texts;
            let mut ends: Vec<usize> = vec![0; nn];
            if i < num_threads as usize - 1 {
                ends[0] =
                    (texts[0].len() + (num_threads as usize)) / (num_threads as usize) * (i + 1);
                let end_seq = &texts[0][table_load_disk(&mut tables[0], ends[0], ratio as usize)..];

                for j in 1..ends.len() {
                    ends[j] = off_disk_position(&texts[j], &mut tables[j], end_seq, ratio as usize);
                }
            } else {
                for j in 0..ends.len() {
                    ends[j] = texts[j].len();
                }
            }

            for j in 0..ends.len() {
                let l = &texts[j][table_load_disk(&mut tables[j], starts[j], ratio as usize)..];
                let l = &l[..std::cmp::min(l.len(), 20)];
                println!("Text{} {:?}", j, l);
            }

            println!("Spawn {}: {:?} {:?}", i, starts, ends);

            let starts2 = starts.clone();
            let ends2 = ends.clone();
            //println!("OK {} {}", starts2, ends2);
            let texts_len2 = texts_len.clone();
            let _one_result = scope.spawn(move || {
                worker(
                    texts,
                    starts2,
                    ends2,
                    texts_len2,
                    i,
                    (*output_file).clone(),
                    (*data_files).clone(),
                    ratio as usize,
                    big_ratio as usize,
                );
            });

            for j in 0..ends.len() {
                starts[j] = ends[j];
            }
        }
    });

    println!("Finish writing");
    let mut buffer = File::create(output_file)?;
    for i in 0..texts.len() - 1 {
        buffer.write_all(&texts[i][..texts[i].len() - HACKSIZE])?;
    }
    buffer.write_all(&texts[texts.len() - 1])?;
    Ok(())
}

fn main() -> std::io::Result<()> {
    let args = Args::parse();

    match &args.command {
        Commands::Make { data_file } => {
            cmd_make(data_file)?;
        }

        Commands::MakePart {
            data_file,
            start_byte,
            end_byte,
        } => {
            cmd_make_part(data_file, *start_byte as u64, *end_byte as u64)?;
        }

        Commands::Merge {
            suffix_path,
            output_file,
            num_threads,
        } => {
            cmd_merge(suffix_path, output_file, *num_threads)?;
        }
    }

    Ok(())
}
