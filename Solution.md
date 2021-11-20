# Glassnode Challange Solution

## Decisions 

### Complex DB query vs. Processing the Data in Go  

I've decided to rather write a more complex DB query than processing the data in the Go application. I think this way the application can focus on the API logic instead of logic to process the data. Additionally most SQL databases optimize queries in terms of of parallelization and frequency of data access and you don't have to load all the data from the database in memory of the application to process it (data locality).

### Docker Image

I choose to use the official Docker Image for Go "golang:1.17.3". This was the easiest and fastes way to get the web service running without having to configure anything. For a production environment that might not be the best fit. Read more about my thoughts regarding this Topic in the [Additional Consideration Section](#additional-considerations).

### Configuration with Env Variables

The API services database connection is configurable with Env variables. This way it is easy to configure the service in differnt environment like container orchestration platforms and tools. This is one of the principles described in the 12 factor apps [12factor][]

### DB Interface and handler struct

The DB is implemented as an Interface in this code because that makes it easy to replace the implementation that connects to a SQL database by another implementation for example for unit testing purposes.

The HTTP handler for the API is defined as a method on a struct called "handler". This way the handler struct can hold a reference to the current database implementation and use it in the handler. There are multiple ways to pass arguments like this DB reference to http.Handler (global variable, wrapped http handler, or struct implementing [http.Handler][]). In fact I've once written a [blog post][] about that. I've decided to use the struct approach because it makes it easy to extend the arguments that should be passed to the handler afterwards.

### SQL Driver

In order to connect to a PostgreSQL database you have to import an appropriate driver I decided to use: [github.com/lib/pq][]. I've used in the past and it has a license allowing free use for any purpose.

## Additional Considerations

### Docker Image Size and security

I used a docker base image of golang, cause this was the fastest way for me to get it running. This base image is larger then it needs to be and the configuration can also be considered as insecure as the process is running with the root user. In a production docker image one should consider using an image that is as small as possible and implement additional security aspects like a dedicated user for the service.

### All code in one package / Better package management

In this project all code is written in a single file and in a single Go package. Once the project grows it should be split into different Files/Packages to keep it well organized.

### Automated Integration Tests / Test with DB integration

I wrote some Unit Tests for the HTTP handler in this web service. But I skipped testing the DB connection. There are multiple ways of testing that.
    
- Unit test SQL commands
    
    - I don't like this approach because every change in the SQL would break the test, and the test won't assure you that your query is correct
 
- Start a DB and run automated integration tests against the DB component
    
    - This is better because it ensures the database will accept the query and responds with the queried data

- Start the whole system (Database & API) and run automated e2e test against the API

    - Even more initial effort compared to the integration test, but this test will assure the overall system is working like expected

### User friendly API response and errors

I did not put much thought in the API responses a user gets when something went wrong. Sometimes it can be usefull to provide additional information like action a user could take to mitigate the issue. Also sometimes you don't want the user to know what exacly went wrong especially on internal server errors as this might give attackers a hint how the internals of your service look like.

In the task it is defined how the API response for a successfull call should look like. I think it would be better to change the JSON keys to something more meaningful than "t" and "v".


[12factor]: https://12factor.net/
[github.com/lib/pq]: https://github.com/lib/pq
[blog post]: https://mj-go.in/golang/pass-arguments-to-http-handlers-in-go
[http.Handler]: https://pkg.go.dev/net/http#Handler