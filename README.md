# src
This is a Go Project which can establish a web system where users can post and check message on it.

S1 : front-end server
S2 : back-end server

S1 listens on port 8080 ( or any other you want ), waiting for connections from browser, remotely calling functions from S2 and then return the page back on the browser.
S2 registers JSON-RPC service on port 1234, every RPC calling is from S1 to S2. 

The package for mysql and ZooKeeper are renamed in my code.

###############################################################

Every function is Browser -> FrontEnd(s1) -> BackEnd(s2) -> FrondEnd(s1) -> Browser.

###############################################################

To manage mysql by Go, we need to import the go-mysql driver, which can be found below:

import "database/sql"
import _ "github.com/go-sql-driver/mysql"


Register: S1 receives request from users and then hands them to s2, then s2 talks to mysql database and check if there is already one username existing. If so, return "Failed", if not, insert it into the database and return "successful".

Login : S1 receives request from users and then hands them to s2, then s2 talks to mysql database and check if the username and password in the request equals to what the database stores. If login successfully, S1 will start a session, keep the sid in the memory. Browser will receive the reply from s1 which can set the cookie on the browser.

Post : Like Register procedure. If the user has already logined, then s1 will get the sid from the cookie in the browser and then get the username. If not, the username will be "Guest" by default. Then the username will also be sent to s2 as well as the message.

Check : Retrive all the data from the table "content" which contains all the posted content.

Logout : Destroy the session which is using now.



################################################################

Run several servers to keep the system stable. Use ZooKeeper to manage the cluster.
If the running server crashes, another leader(running server) will be elected from the other servers by ZooKeeper.

To use ZooKeeper, first we need to install and start the ZooKeeper, then connect our server to ZooKeeper.
Since ZK only provides API for C and Java, so we need import API for Go manually.

import "github.com/samuel/go-zookeeper/zk"

