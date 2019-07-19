# About Project

This is a project created in Go lang specially because of the fast processing and performance
It can handle simultaneously API request and execute them one by one and return the status on the server.
Persistent Queue is used to log the Api requests and return the status.

# How It Works

A server is created, having two paths, one for logging API requests in Queue and other for getting the status of the request

* http://localhost:8080/validatejob?id=6

This logs the request in Persistent queue with id

* http://localhost:8080/getjobstatus?id=6

This returns the status of the API Request with the mentioned id `Processing` or `Done`

# Note

~ This was created as a freelancing project

## Created By ~Prakhar