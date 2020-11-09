# coding-challenge

## Purpose
This project implements a GraphQL based microservice written entirely in Go. The purpose of this microservice is to perform DNS lookups of IPV4 IP addreses to determine if the IP addresses are on a block list. This service is useful for gathering threat intelligence and can be used, for example, by mail servers to block emails from being processed if the email sender's IP address is on a block list.

## Implementation
This microservice is implemented in Golang and uses the gqlgen package as a framework for implementing the GraphQL interface. Sqlite3 is used as the backend database for storing the requested DNS blocklist details for each of the IP addreses provided via the GraphQL interface. Currently blocklist information is only collected for IPV4 addresses. 

## Job Queue
This microservice also implements a job queue using Golang channel serviced by an asyncronous go routine that is used for collecting DNS blocklist information for each IP address. 

### Authentication ###
Basic authentication is also implemented to protect the primary GraphQL interface by only allowing authenticated users to access the system. Currently the only user that is able to access the system is the user with the following credentials:
- **Username** : secureworks
- **Password** : supersecret

Authentication is implemented via a GraphQL **authenticate** mutation that accepts username and password as input and generates a JWT bearer token if the username and password have been sucessfully authenticated. The JWT bearer token is then expected to be used in all other GraphQL API queries and mutations by supplying the JWT bearer token as the value of an HTTP **Authorization Header**. If the HTTP **Authorization Header** is not supplied on any other GraphQL API then the API call will fail. 

The microservice makes use of an HTTP Authentication middleware handler that wraps each of the GraphQL handlers/resolvers that verifies the presence of JWT token and performs the appropriate validation. Validation includes checking the username and password claims in the JWT token to make sure that user is still valid.

### GraphQL API
Besides the authenticat mutation the GraphQL interface provides 2 primary enty points:

- **enqueue** - mutation for asyncrhonously collecting DNS blocklist details for one or more IPV4 addresses
- **getIPDetails** - query for obtaining blocklist details for a single IPV4 address

### Packages
The implementation of this microservice depends on the following 3rd party Golang packages to provide its capabilities:
- **[gqlgen](https://github.com/99designs/gqlgen)**  - Framework to generate Golang code in order build to implement graphql servers
-	**[sqlite3](https://github.com/mattn/go-sqlite3)**  - sqlite3 database driver
- **[godnsbl](https://github.com/nerdbaggy/godnsbl)** - DNS Blocklist lookup functionality
- **[jwt-go](https://github.com/dgrijalva/jwt-go)**  - Library for creating and validating JWTs used by this application to provide authentication
- **[chi](https://github.com/go-chi/chi)**        - HTTP router which provides support for HTTP middleware, specifically authentication middleware
-	**[uuid](https://github.com/google/uuid)**       - Library for generating [RFC 4122](http://tools.ietf.org/html/rfc4122) UUIDs 
-	**[testify](https://github.com/stretchr/testify)**  - Tools for testifying that your code will behave as you intend
