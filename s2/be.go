  package main

  import (
	  "database/sql"
	  "fmt"
	  "log"
	  _ "mysql-master/mysql"
	  "net"
	  "net/rpc"
	  "net/rpc/jsonrpc"
	  _ "session/providers/memory"
  )

var DB = &sql.DB{}
var contDB = &sql.DB{}
// initialize the connection to the database
// gouqizi123 is the password, followed by address, port, and database name.
func Init(){
    DB,_ = sql.Open("mysql","root:gouqizi123@tcp(127.0.0.1:3306)/account")
	contDB,_ = sql.Open("mysql","root:gouqizi123@tcp(127.0.0.1:3306)/mysql")
}

// Req type is for registering the RPC servce.
type Req struct{
}

  const lenPath = len("/view/")

  // content from the browser.
  type Page struct {
        Title   string
       	Body    []byte
  }
  type Pass struct {
	Username string
	Page   Page
  }
  // account info
  type Info struct {
	Username string
	Password string
  }

  // API for register
  func (r *Req) Create(i Info, reply *bool) error {
    rows, _ := DB.Query("SELECT * FROM loginuser where username = ?",i.Username)
    var un string
    var pw string
    for rows.Next(){
    	if err := rows.Scan(&un,&pw); err == nil {
    		fmt.Println("This username has already been there!!!")
    		*reply = false
    		return nil
    	}
    }
	DB.Exec(
		"INSERT INTO loginuser (username, password) values (?,?)",
		i.Username,i.Password)
      fmt.Println("successfully")
	*reply = true
	return nil
  }

  // API for Login
  func (r *Req) Login(i Info, reply *bool) error {
    rows, _ := DB.Query("SELECT * FROM loginuser where username = ?",i.Username)
    var un string
    var pw string
    for rows.Next(){
        if err := rows.Scan(&un,&pw); err != nil {
            log.Fatal(err)
        }
    }
	if i.Password == pw {
		*reply = true
	}else   {
		*reply = false
	}
	return nil
  }

  // API for posting content
  func (r *Req) Post(p Pass, reply *bool) error {
	  contDB.Exec(
		  "INSERT INTO content (user, title, content) values (?,?,?)",
		  p.Username,
		  p.Page.Title,
		  p.Page.Body)
	*reply = true
        return nil
  }

  // API for check the content
  func (r *Req) Get(usern string, reply *[]byte) error {
	  rows, _ := contDB.Query("SELECT * FROM content")
	  var tempres = ""
	  var un string
	  var tl string
	  var bd string
	  for rows.Next(){

		  if err := rows.Scan(&un,&tl,&bd); err != nil {
			  log.Fatal(err)
		  }
		  //fmt.Printf("name:%s ,id:is %d\n", name, id)
		  tempres = tempres + un + ":" + tl + " " + bd + "\n"
	  }

	//for u,v := range cont {
	//	tempres = tempres + u + ":" + v.Title+" " + string(v.Body) + "\n"
	//}
	*reply = []byte(tempres)
	return nil
  }
//  func (r *Req) Create(args Page, reply *int) error {
//	temp := &page{title: args.Title, body:[]byte(args.Body)}
//	
//		args.save()
//		*reply = 1
//		return nil	
//	}

//   func (r *Req) Login(args string, reply *bool) error {
//
//        b, _ := (ioutil.ReadFile("account.txt"))
//        b2 := string(b[:])
//        if b2 == args {
//                *reply = true
//        }else {
//                *reply = false
//        }
//	return nil
//  }

  func main() {

	  Init();

	  // configuration of ZooKeeper
	  zkConfig := &ZookeeperConfig{
		  Servers:    []string{"localhost"},
		  RootPath:   "/ElectMasterDemo",
		  MasterPath: "/master",
	  }
	  // channel for the electing result
	  isMasterChan := make(chan bool)

	  var isMaster bool

	  // run the election
	  electionManager := NewElectionManager(zkConfig, isMasterChan)
	  go electionManager.Run()

	  for {
		  select {
		  case isMaster = <-isMasterChan:
			  if isMaster {
				  // do jobs on master
				  rpc.Register(new(Req))
				  //rpc.HandleHTTP()
				  l, e := net.Listen("tcp", ":1234")
				  if e != nil {
					  log.Fatal("listen error:", e)
				  }
				  for {
					  conn, err := l.Accept()
					  if err != nil {
						  continue
					  }
					  go jsonrpc.ServeConn(conn)
				  }
			  }
		  }
	  }



  }
