# GriddyGo

## Example API Calls
1. GET: <IP>:3000/data
    - Returns all data

2. GET: <IP>:3000/data?key=id
    - Returns the value connected to the key + all the 
    values connected to the foreign key

3. GET: <IP>:3000/data use the Header, key: "key", value: id
    - Returns the value connected to the key + all the 
    values connected to the foreign key

##### POST Structure
{
  "value" - string - required,
  "t1key" - int - optional
}

4. POST: <IP>:3000/data use the Body, 
{
  "value": "Example Text"
}
    - Returns msg: "Successfully added 'Example Text' with Key of: x into T1"

5. POST: <IP>:3000/data use the Body, 
{
  "value": "Example Text"
}
    - Returns msg: "That value is already taken in the database, please insert a unique value."

6. POST: <IP>:3000/data use the Body, 
{
  "value": "Example Text with Foreign Key",
  "t1key": 2
}
    - Returns msg: "Successfully added 'Example Text' with Key of: x into T2"

##### DELETE Structure
Only one is required
{
  "key" - int - represents the t1 key column
  "t2key" - int - represents the t2 key column
}

7. DELETE <IP>:3000/data use the Body, 
{
  "key": 6
}
    - Returns msg: "Key deleted successfully from table T1"

## Task
1. GitHub –

* Setup free public repo on GitHub –

2. Golang –

* Create a Golang based web-service with the following functions –

  * Create a route for –

    * GET /data –

      * Return an empty string if no data posted –

    * POST /data –

      * Post a string (used on GET request) –

    * DELETE /data –

      * Set string back to empty –

3. AWS –

  * Setup free AWS account –

  * Create EC2 server –

  * Migrate your code to the EC2 server –

  * Deploy code to ec2 server (aka via Git) –

  * Open firewall –

    * Only open the firewall when you’re ready to share (let me know via email) –

4. Database –

  * Spin up RDS Postgres instance –

    * We will store data now versus using a variable in Go –

  * Create tables named T1, T2 with following schemas –

    * T1 (PK int key, unique string value) –

    * T2 (PK int key, FK int t1key, string value, default date) –

  * Install postgres client on ec2 instance –

  * Generate data for T2 and populate (can be random) –

  * Write query to join data (inner, left, right) –

  * Send queries to join table on T1.key and T2.t1key to me –

5. Combine –

  * Modify routes in your Golang server to store and retrieve data from database –

  * The data retrieval should be a join of T1 and T2 –

6. Explain –

7. Discuss the technologies involved –

8. How might it be improved –

  * Security –

  * Monitoring –

  * Logging –

9. Optional –

  * Implement HTTPS –

  * Implement gRPC instead or in-addition of REST
