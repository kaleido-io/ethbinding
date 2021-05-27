# ethbinding

A set of type and function definitions, building a dynamic linking library to LGPL code in the go-ethereum repository.

The `ethbinding.EthAPIShim` interface is exposed by the `EthAPIShim` shim built from this dynamic link library.
That interface can be used by projects as headers to bind to this library, without pulling in the go-ethereum implementations directly into their binary distributions.


