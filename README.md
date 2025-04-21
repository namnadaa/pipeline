# Go Data Pipeline

A simple data processing pipeline written in Go.  
This project demonstrates a modular architecture, channel-based filtering, and a custom ring buffer with timed flushing.

## Features

- Multi-stage data filtering
- Removes negative and non-relevant values
- Buffering with a custom ring buffer
- Concurrent processing using goroutines and channels
- Reads input data from the console
  
## How It Works

The pipeline reads integers from the console, processes them through multiple stages, and outputs the results.

### Processing Stages:

1. **NegativeFilterStage**  
   Filters out all **negative** numbers.

2. **NotDividedFilterStage**  
   Filters out numbers **not divisible by 3**, including **0**.

3. **BufferStage**  
   Buffers the remaining numbers using a **ring buffer** implemented with a linked list.  
   The buffer is automatically flushed at regular time intervals and sends the data forward.

### Ring Buffer

- Manually implemented using linked list `Node` structures.
- Configurable via constants:
  - `bufferSize` — maximum number of elements in the buffer.
  - `bufferDrainInterval` — how often to flush the buffer (e.g., every 10 seconds).

### Input Source

- Data is entered **manually through the console**.
- To stop the input — type `exit`.
