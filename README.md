[![Build Status](https://travis-ci.org/c4dt/drynx.svg?branch=master)](https://travis-ci.org/c4dt/drynx)
[![Go Report Card](https://goreportcard.com/badge/github.com/c4dt/drynx)](https://goreportcard.com/report/github.com/c4dt/drynx)
[![Coverage Status](https://coveralls.io/repos/github/c4dt/drynx/badge.svg?branch=master)](https://coveralls.io/github/c4dt/drynx?branch=master)

# Drynx

Drynx is a library for simulating a privacy-preserving and verifiable data sharing/querying tool. It offers a series of independent protocols that when combined offer a verifiably-secure and safe way to compute statistics and train basic machine learning models on distributed sensitive data (e.g., medical data).

This is a [fork of LDS drynx](https://github.com/ldsec/drynx), this is what is officially running at [C4DT](https://c4dt.org). It has some stabilising features that are being merged upstream:

 * datasets loaders, found in `lib/provider/loaders`
   * one keeping the upstream behavior of randomly generating data
   * one file based, reading CSV values and populating the Data Provider with it
 * CLI interface, found in `cmd`
   * `server` to run a Drynx node
   * `client` to communicate from the command line with the nodes
 * some others changes
   * allow to select which column to query
   * splits of some structures to allow for dedis/protobuf auto-generation
   * docker image for the server
   * some revamp of operation to more modularise them
