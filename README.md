# coding-challenge

## Purpose
This project implements a **GraphQL** based microservice written entirely in Go. The purpose of this microservice is to perform DNS lookups of IPV4 IP addreses to determine if the IP addresses are on a block list. This service is useful for gathering threat intelligence and can be used, for example, by mail servers to block emails from being processed if the email sender's IP address is on a block list.

## Implementation
This microservice is implemented in Golang and uses the **[gqlgen](https://github.com/99designs/gqlgen)**  package as a framework for implementing the GraphQL interface. The microservice also uses a backend database for storing the requested DNS blocklist details for each of the IP addreses provided via the GraphQL interface. Currently blocklist information is only collected for IPV4 addresses. 

### Packages
The implementation of this microservice depends on the following 3rd party Golang packages to provide its capabilities. These packages are managed via **go mod**  and are maintained in the **go.mod** file. The following are the primary packages used by this microservice:

- **[gqlgen](https://github.com/99designs/gqlgen)**  - Framework to generate Golang code in order build and implement graphql servers
- **[sqlite3](https://github.com/mattn/go-sqlite3)**  - sqlite3 database driver
- **[godnsbl](https://github.com/nerdbaggy/godnsbl)** - DNS Blocklist lookup functionality
- **[jwt-go](https://github.com/dgrijalva/jwt-go)**  - Library for creating and validating JWTs used by this application to provide authentication
- **[chi](https://github.com/go-chi/chi)**        - HTTP router which provides support for HTTP middleware, specifically authentication middleware
-	**[uuid](https://github.com/google/uuid)**       - Library for generating [RFC 4122](http://tools.ietf.org/html/rfc4122) UUIDs 
-	**[testify](https://github.com/stretchr/testify)**  - Tools for testifying that your code will behave as you intend

### Configuration
A configuration file in JSON syntax is used to specify various configuration file options. Here is the supplied **config.json** configuration file:

```
{ 
    "server" : {
        "listening_port": 8080
    },
    "logger": {
        "log_file_name": "./coding_challenge.log"
    },
    "db": {
        "db_type": "sqlite3",
        "db_path": "./coding_challenge.db",
        "persist": true
    },
    "dnsbl": {
        "blocklist_domains": [
            "zen.spamhaus.org"
        ]
    },
    "auth" : {
        "username": "secureworks",
        "password": "supersecret",
        "expiration_duration": 15
    },
    "job_queue": {
        "queue_length": 100
    }
}
```

### Logging
The microservice generates log messages to a log file using the standard golang "log" package. The log file name is **coding_challenge.log** and can be changed by modifying the **log_file_name** attribute in the **logger** section of the **config.json** file.

This log file can be used for debugging and informational purposes.

### DNSBL
IP Addresses can be checked against one or more blocklist domains. As per the requirements of this coding challenge, the blocklist domain being used is **[zen.spamhaus.org](https://www.spamhaus.org/faq/section/DNSBL%20Usage)** as configured in the **blocklist_domains** attribute of the **dnsbl** section of the **config.json** file.

### Database
The database used for storing the blocklist details is an sqlite3 database in a database file named **coding_challenge.db**. The file name of the database can be changed by specifying a new dabase file name in the **db_path** attribute of the **db** section of the **config.json** file. 

Additionally, by default, the datbase is persisted across various instantiations of the microservice. If the database is to not be persisted, the default behavior can be changed by setting the **persist** attribute to **false** in the **db** section of the **config.json** file.

### Job Queue
This microservice also implements a job queue using a Golang channel serviced by an asyncronous go routine that is used for collecting DNS blocklist information for each IP address. The default queue length is 100, but can be overridden by setting a new value in the **queue_length** attribute of the **job_queue** section of the **config.json** file.

### Authentication
Basic authentication is also implemented to protect the primary GraphQL interface by only allowing authenticated users to access the API. Currently the only user that will be authenticated to use the API is the user with the following credentials:
- **Username** : secureworks
- **Password** : supersecret

Authentication is implemented via a GraphQL **authenticate** mutation that accepts an username and password as input and generates a JWT bearer token if the username and password have been sucessfully authenticated. The JWT bearer token is then expected to be used in all other GraphQL API queries and mutations by supplying the JWT bearer token as the value of an HTTP **Authorization Header**. If the HTTP **Authorization Header** is not supplied on any other GraphQL API call then the API call will fail. 

The microservice makes use of an HTTP Authentication middleware handler that wraps each of the GraphQL handlers/resolvers that verifies the presence of JWT token and performs appropriate validation. Validation includes checking the username and password claims in the JWT token to make sure that user is still valid. 

Additionaly validation includes checking to see if the token has expired. Currently the token has a default expiration duration of 15 mintues. The default expiration duration can be changed by modifying the **expiration_duration** attribute of the **auth** section of the **config.json** file.

### GraphQL API
The GraphQL API is served by default on port **8080**, but the port can be configued by changing the **listening_port** attribute in the **server** section of the **config.json** configuration file.

Besides the **authenticate** mutation the GraphQL interface provides 2 primary end points:

- **enqueue** - mutation to asyncrhonously queue a job to the job queue to collect DNS blocklist details for one or more IPV4 addresses. 
- **getIPDetails** - query for obtaining blocklist details for a single IPV4 address. This returns a DNSBlocklistRecord which contains a response_code
                     field providing blocklist information about the IPV4 address. Detailed information about the response_code values can be found at                                          **[zen.spamhaus.org](https://www.spamhaus.org/faq/section/DNSBL%20Usage#200)**

### GraphQL Schema
The following is the GraphQL schema implemented by this microservice:
```
"""
Timestamp data type
"""
scalar Time

"""
Returned by the authenticate mutation.
Content of bearer_token is required to be supplied in 
Authorization Header on all subsequent API calls.
"""
type AuthToken {
  """
  bearer_token contins JWT token string
  """
  bearer_token: String!
}

"""
Contains information about whether or not an IPV4 address is on a blocklist
"""
type DNSBlockListRecord {
  """
  A unique identifier generated by the system for each record
  """
  uuid: ID!

  """
  Timestamp indicating when the record was first created
  """
  created_at: Time!

  """
  Timestamp indicating when the record was last updated
  """
  updated_at: Time!

  """
  Indicates if the ip_address is on the blocklist. For detailed information on the response code values, 
  you can refer to https://www.spamhaus.org/faq/section/DNSBL%20Usage#200
  """
  response_code: String!

  """
  IPV4 address of the record
  """
  ip_address: String!
}

"""
Coding Challenge Queries
"""
type Query {
  """
  Provides DNS blocklist information for the specified IPV4 address. If the ip address has not been previously specified
  in a previous enqueue mutation, then a DNSBlockListRecord will be returned with an empty uuid and a response_code of "NXDOMAIN"
  """
  getIPDetails(ip: String): DNSBlockListRecord
}

"""
Coding Challenge Mutations
"""
type Mutation {
  """
  Used to autenticate the supplied username and password and to return and AuthToken to be used on subsequent API calls
  """
  authenticate(username: String!, password: String!): AuthToken!

  """
  Used to queue an array of IPV4 addresses onto the aysnchronous job queue so that blocklist information can be obtained
  for those IP Addresses. If the queue is full, then an error will be returned indicating the queue is currently full, and that a retry
  should be attempted.
  """
  enqueue(ip: [String!]!): Boolean
}
```

## Tests
A set of system level tests have been implemented to perform postive and negative testing of the GraphQL interface. Additionally, unit tests have been written to test the supporting packages.

## How to Install
This package can be installed with the go get command:

    go get github.com/egreen64/coding-challenge
    
## How to Build and Test
There are 2 ways that this project can be built and tests can be run: 
- Natively using the go compiler installed on the same machine you installed the package
- Locally using a golang docker build container.

### Native Build and Test
In order to do a native build, you need to make sure that you have installed a golang compiler on the same machine you installed this package. This package has been built using version 1.15.3 and it is recommended that this version of the compiler be installed.

This package can be built with the go command:

    go build
    
This package can be tested with the go command:

    go test -v ./...   
    
### Local Build and Test using Docker
In order to do a local build and test, you need to make sure that you have Docker installed on the same machine you installed this package. 

This package can be built using the following supplied script:

    ./build.sh
    
This script will build the package, run the test scripts and then create a Docker image named **coding_challenge:latest**

Additionally the **build.sh** script can push the image to Docker hub.

You will need to uncomment the last 2 lines of the **build.sh** script to perform the docker push as well as change the repository name from **egreen6464** to your own repoistory name:
    
    #push docker image
    #docker tag coding_challenge:latest egreen6464/coding_challenge
    #docker push egreen6464/coding_challenge:latest

## How to Run
Just as this package can be built and tested either natively or locally, the package can also be run either natively or locally.

### Running Natively

This packge can be run natively with the following command:

    ./codingchallenge
    
### Running Locally using Docker

This package can be run locally with the following supplied script:

    ./run.sh [port]
   
where **port** is the port number that can be specifieid to override the default listening port of **8080**

This script uses Docker to run the container **coding_challenge:latest** that was created previously by the local build.

## Helm Chart

This project also includes a Helm chart for deploying the docker image **egreen6464/coding_challenge:latest**, which is created during the local build, to a k8s cluster. The helm chart for this project was created with Helm V3. The helm chart can be found in the project's **/helm** directory.

To deploy the helm chart, use the following helm CLI command from a machine where helm is installed:
    
    helm install codingchallenge ./helm

To establish connectivity to the container, you need to execute the commands in the Notes that are displayed after sucessfully executing the above command. Here are the commands that should be displayed in the Notes:

    export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=codingchallenge,app.kubernetes.io/instance=codingchallenge" -o jsonpath="{.items[0].metadata.name}")
    export CONTAINER_PORT=$(kubectl get pod --namespace default $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
    echo "Visit http://127.0.0.1:8080 to use your application"
    kubectl --namespace default port-forward $POD_NAME 8080:$CONTAINER_PORT
    
# Have fun and enjoy!
  
