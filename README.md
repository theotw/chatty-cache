# Chatty Cache
This is a simple caching system that holds values in memory
using a last recently used purging system.
Cache size is bounded to specific number of bytes set by the user at startup.

Optionally, cache messages can be shared across nodes in a system for an in memory distributed cache.

Right now it uses NATS as its method to distribute cache data