This is a starting of a port of the libdrynx-crypto parts. It implements the
encryption and decryption of an integer.

TODO:
- clean up the crypto-browserify @ts-ignore
- put the methods in its own file and only export the classes in index
- re-arrange classes and methods to fit Valérian's likings ;)
- check with DEDIS why `new BN256Scalar(randomBytes(32))` sometimes fails to produce a good point - see KeyPair.create
- add test-vectors taken from running encryption in libdrynx

LATER OPTIMIZATIONS (once the UI works):
- implement the TODOs in pointToInt
