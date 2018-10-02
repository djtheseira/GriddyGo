## Notes
##### Simple Compile/Run Instructions
Compile the file with go build in the folder directory
Run compiled file with ./[filename]

##### Go Notes
* To export a variable, the first letter must be capital. So for pi, use "math.Pi"

* Arrays are cool, but slicing is more popular.
  * When slicing, any changes impact the underlying array.

* Pointers: &, \*

##### Go?
* Golang is: 
  1. Compiled
  2. Concurrent
  3. Statically-typed
  4. Garbage-collected
  5. Efficient
  6. Scalable
  7. Readable
  8. Fast
  9. Open-source

* Built to run quickly with concurrency, doesn't run in threads. 

* The "main" package is how everything is the core of how everything is run.

## Explain the Process
The process of creating this application started with first learning the basics of 
Golang, going through the documentation online and getting a brief, high-level 
understanding of the language. After that, I started writing tiny bits of code to
get the app started and running with just a string value. Once I was able to get 
that part down I implemented the basic DELETE, GET, POST methods that revolved
around a simple string variable. Once I felt comfortable enough with the basics,
came the tricky part, getting all the different parts of AWS connected and working.

As it was my first real attempt at putting multiple parts of AWS to use, I used a
few tutorials I found online to grasp some understanding of how to get the ball
rolling. First, I created the EC2 instance with linux, then the Postgres DB. 
After this, I got "stuck" on how to get the code onto the server in an "continuous"
way. So I was reading how to use the Code Pipeline(C.P.) and Code Deploy(C.D.) services and at
first I was able to get the Code Deploy service up and running with only a few 
hiccups. I wasn't satisfied with the constant necessity to push the code up myself
using Code Deploy, so I decided to try out Code Pipeline. After already having 
C.D. setup, it was fairly straight forward (based on the tutorial I was reading)
to get C.P. setup without any sort of validation and testing.

## Discuss the Technologies Used
Golang - A backend language to simplify how server-side software is created. It's run
  concurrently, not depending on threads, providing a pretty simple way to take 
  advantage of multiprocessors. Coming from a predominantly Java focues developer, 
  the concepts behind Golang were a little confusing at first. Upon further
  studying and "Hello, World"-ing some test examples, I was able to get a novice
  understanding of the language and how things operate. I would definitely like to
  see how this code works on a much larger scale and how it operates under heavy loads.

PostgreSQL - A type of SQL database, I've been using it the past few years at my 
  current position. It's a very powerful type of database that allows for
  fantastic ways to manipulate geospatial data natively.

AWS EC2 - The service that hosts the linux based server to which the Go app is 
  served. First time really getting my hands dirty using AWS, but setting up the
  EC2 instance was pretty simple. Since I've setup my own linux server that hosts 
  my personal website, getting around the command line with basic commands 
  wasn't too bad.

AWS RDS - The service that hosts the relational database, in this case Postgres. 
  Just as straight forward as EC2, with just a few clicks of the mouse, and 
  changing the database info, the instance was up and running. I used DBeaver as
  my SQL IDE, where I wrote up the basic queries to create the tables and input
  sample data.

AWS Code Deploy - Probably the trickiest part of all this, just due to having 
  to make sure the security settings and all those other things are setup
  correctly. Once this was setup, it was easy to get the code up to the proper
  place in the server without SSH-ing.

AWS Code Pipeline - My favorite part because continuous integration is such an
  useful tool to have for developers, tech leads, SA's, or whoever is managing
  the code. The one major thing missing is some sort of testing software with
  something like Jenkins.

## How to Improve the Project

* Security, Monitoring, Logging

From a developer standpoint, one major way to improve this is having basic
unit tests setup. On top of that, implementing integration testing with 
something like Jenkin, to prevent anything from going up to production that
shouldn't be. Properly checking and logging errors into a table that helps
to notify the Developers what's going on and where to search for the issue.
Ensuring that if anything breaks, it won't take the entire server down and 
prevent anyone from accessing the API.

From a security standpoint, the first thing I can think of that most API's
have, is some sort of auth key, thats provided by the user to help track
down who's using the API. Also, it helps to authorize only certain groups
to use certain parts of the API, such as: DELETE or POST. Monitoring the 
servers with some sort of uptime robot that notifies the SA when something
is wrong and needs attention, as well as implementing some sort of 
autoscaling mechanism for times when the load gets to be too heavy, so that
the API experiences little to no downtime.