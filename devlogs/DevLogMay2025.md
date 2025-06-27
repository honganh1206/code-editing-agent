The first dev log! The idea of clue came from [this wonderful on how to build an Agent by Thorsten Ball](https://ampcode.com/how-to-build-an-agent)

So what so special about clue? Hard to answer, since clue is pretty run-off-the-mill: jan, cline, serena... there are probably hundreds of them out there with thousand of users daily. So my answer would be:

1. It's written in Go(not really special, I know) and
2. It's CLI-based (booooo think of better things to say)

That is why I do not intend to commercialize it, and clue would stay to be a for-fun thing (for now, contact me VCs if you see something potential)

But I have a handful of ideas at the moment: Expanding the toolset, integrating different models into it, creating profiles, etc., and for a person who loves tinkering like me, this is like a kid walking into a candy store. I love having ideas for my work and try to make them come to live and become useful.

What I am hacking on is an interface for different LLMs. The most challenging thing is that dealing with the structure for the requests/responses. I'm using the [anthropic-go-sdk from the Stainless API team](https://github.com/anthropics/anthropic-sdk-go) and while working on Claude alone is a thing, trying to unify the structures for different models is another thing, and apparently not a small feat. But he staying-up-late hours is quite worth it, and now the interface is working quite well.

The key is to know how the requests and responses go back and forth. Underlying them are messages, and underneath the messages are blocks of content, with each block handled with care :) Each block has a type, and each type must go with either the user and assistant, so 90% of the time I was working on the interface was spent on debugging the types and how messages are lined up. Having done with that, I gained so much knowledge about how LLMs work.

The guys at Anthropic gave [the best explanation](https://docs.anthropic.com/en/docs/build-with-claude/tool-use/overview#how-tool-use-works) right here

And shoutout to the dudes working on the SDK at Stainless! Your work is stellar!

That's all for the 1st dev log! Thank you for reading :)

See you next dev log!

May 8 2025
