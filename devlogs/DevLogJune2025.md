[Delta vs snapshot streaming](https://docs.anthropic.com/en/docs/build-with-claude/streaming#delta-vs-snapshot-streaming)

Maybe a buffer is enoug? Get the stream, push the data from the stream to the buffer, then send the data from the buffer to a channel of CustomResponse? -> Go for this, dont overthink

content_block_start events for tool use blocks
content_block_delta events with accumulated partial JSON

`ContentBlock` as a unified interface for different content block types

sqlite3 as a lightweight option to store conversations

Two modes: Snapshot and streaming

Too many tools and the agent would stuck and not know which one to use. A curated set of tools is important

The tokens must flow. The agent should retry the operation instead of halting it

Next tools:

Structural Search Interface (full‑text, regex, and language-aware structural code search)?
Commit Diff Lookup

Server to handle CRUD for conversations, lifecycle for server

No need for store, that's for local model state
