# grpcstuff consistent hashing 'poc'

This is an implementation of a simple event sourced microservice that looks up a customer's
current state.

this is a sketch of using Gossip, Consistent hashing and a super fast kv store to build
a bespoke eventsourced datastore.  That said: its Not really a full PoC; but it covers
the majority of what is needed.


> The code is heavilly commented; including lots of 'potential enhancment here'

> Testing: I have basic unit tests for the data package.

## Why

based  on a conversation; and because its _supprisingly_ easy; and it made more sense in
the historical context than now.  this is because a few extra tools mean its probably easier
to just use ristretto than do it this way.  though combining the approaches would have merit (see below
and comments in service.customer.go)

Its missing a lot of validation and needed things, like rebalancing, if it is to work in this format.
This is stuff that would have to be built in; but the needed pieces are already there.

1) consistent hashing with replicaiton supports rebalancing based on load; you'd need to implement
   the transport layer
2) gossip manages the server member list and the health.  its evenetually consistent and also
   gives you a inter-server communicaiton layer to manage things like rebalancing, adding and removing
   nodes.
3) backup/restore is supported by badger.  Streaming is supported there
4) i leave a huge amount of performance on the table here because of time constraints.

Load balancing can be done client side; using something like envoy, istio or even just basic round
robin addressing the members by looking up which member owns a partion using the consistent hashing library.
basically, wrap up the grpc client with a simple lookup layer using the consistent hashing and memberlist



## Alternatives

now that ristretto is so easy (and maybe even as a easier implementation):

swap out the badgerdb data.Storer with one that uses ristretto + makes a call to a slower storage
engine.  this way you still get to take advantage of a denser keyspace for cache lookups -> less cache misses

and you don't need to do all the crazy self managment.


## other thoughts.


