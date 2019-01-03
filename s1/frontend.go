  package main

  import (
	  "html/template"
	  "io/ioutil"
	  "log"
	  "net/http"
	  "net/rpc/jsonrpc"
	  "session"
	  _ "session/providers/memory"
  )
var globalSessions *session.Manager
  type Page struct {
	Title string
	Body    []byte
  }
  type Pass struct {
	Username string
	Page    Page
  }
  func (p *Page) save() {
  	filename := p.Title + ".txt"
  	ioutil.WriteFile(filename, p.Body, 0600)
  }

  type Info struct {
        Username string
        Password string
  }

  const lenPath = len("/view/")

   func loadPage(title string) *Page {
  	filename := title + ".txt"
  	body ,_ := ioutil.ReadFile(filename)
  	return &Page{Title: title, Body: body}
  }

  func viewHandler(w http.ResponseWriter, r *http.Request) {
  	title := r.URL.Path[lenPath:]
  	p := loadPage(title)
  	renderTemplate(w, "view", p)
  }

   func saveHandler(w http.ResponseWriter, r *http.Request) {
  	i := &Info{Username: r.FormValue("username"), Password: r.FormValue("password")}
	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
        	log.Fatal("dialing:", err)
	}

  	var args = i
	var reply bool
	err = client.Call("Req.Create", args, &reply)
	if err != nil {
        	log.Fatal("arith error:", err)
	}
  	if reply {
		http.Redirect(w, r, "/success/", http.StatusFound)
	}else{
		http.Redirect(w, r, "/fail/", http.StatusFound)
	}
  }
   func login(w http.ResponseWriter, r *http.Request) {	
//        body := r.FormValue("username")+r.FormValue("password")
	U := r.FormValue("username")
	P := r.FormValue("password")
	i := &Info{Username: U, Password: P}
	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
        	log.Fatal("dialing:", err)
	}

	var args = i
	var reply bool
	err = client.Call("Req.Login",args, &reply)
	if err != nil {
        	log.Fatal("arith error:", err)
	}
	if reply == true {
		//curUser = U
		sess := globalSessions.SessionStart(w, r)
		sess.Set("username", U)
		http.Redirect(w, r, "/view/success/", http.StatusFound)
	}else {
		http.Redirect(w, r, "/view/fail/", http.StatusFound)
	}
  }
  func loginHandler(w http.ResponseWriter, r *http.Request) {
	    sess := globalSessions.SessionStart(w, r)
	    uname := sess.Get("username")
	    if uname == nil {
			globalSessions.SessionDestroy(w, r)
			title := r.URL.Path[lenPath:]
			p := loadPage(title)
			renderTemplate(w, "login", p)
		}else{
		  http.Redirect(w, r, "/view/fail/", http.StatusFound)
	  }
  }

  func logoutHandler(w http.ResponseWriter, r *http.Request) {
	  sess := globalSessions.SessionStart(w, r)
	  uname := sess.Get("username")
	  if uname != nil {
		  globalSessions.SessionDestroy(w, r)
		  http.Redirect(w, r, "/view/success/", http.StatusFound)
	  }else{
		  http.Redirect(w, r, "/view/fail/", http.StatusFound)
	  }
  }
  func createHandler(w http.ResponseWriter, r *http.Request) {
  	title := r.URL.Path[lenPath:]
  	p := loadPage(title)
  	renderTemplate(w, "create", p)
  }
  func succHandler(w http.ResponseWriter, r *http.Request) {
        title := r.URL.Path[lenPath:]
        p := loadPage(title)
        renderTemplate(w, "successful", p)
  }
  func failHandler(w http.ResponseWriter, r *http.Request) {
        title := r.URL.Path[lenPath:]
        p := loadPage(title)
        renderTemplate(w, "failed", p)
  }
  func postHandler(w http.ResponseWriter, r *http.Request) {
        title := r.URL.Path[lenPath:]
        p := loadPage(title)
        renderTemplate(w, "post", p)
  }
  func po(w http.ResponseWriter, r *http.Request) {
//        body := r.FormValue("username")+r.FormValue("password")
        p := &Page{Title: r.FormValue("title"), Body: []byte(r.FormValue("content"))}
        sess := globalSessions.SessionStart(w, r)
	    uname := sess.Get("username")
	    if uname == nil{
	    	uname = "Guest"
		}
		var account string
	    switch v := uname.(type) {
		case string:
			account = v
		}
	    p2 := &Pass{Username: account, Page: *p}
        client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")
        if err != nil {
                log.Fatal("dialing:", err)
        }

        var args = p2
        var reply bool
        err = client.Call("Req.Post",args, &reply)
        if err != nil {
                log.Fatal("arith error:", err)
        }
        if reply == true {
                http.Redirect(w, r, "/view/success/", http.StatusFound)
        }else {
                http.Redirect(w, r, "/view/fail/", http.StatusFound)
        }
  }  
  func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  	t, _ := template.ParseFiles(tmpl+".html")
  	t.Execute(w,p)
  }
  func check(w http.ResponseWriter, r *http.Request)  {
        client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")
        if err != nil {
                log.Fatal("dialing:", err)
        }       
        
        var args = "gggg"
        var reply []byte
        err = client.Call("Req.Get",args, &reply)
        if err != nil {
                log.Fatal("arith error:", err)
        }       
        w.Write(reply)
  }

  func Init() {
	  globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	  go globalSessions.GC()
  }
  func main() {
  	Init()
  	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/view/success/", succHandler)
	http.HandleFunc("/view/fail/", failHandler)
	http.HandleFunc("/create/", createHandler)
	  http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/save/", saveHandler)
        http.HandleFunc("/login/", loginHandler)
        http.HandleFunc("/login/lg", login)
	http.HandleFunc("/post/", postHandler)
	http.HandleFunc("/post/po", po)
	http.HandleFunc("/check/",check)
  	http.ListenAndServe(":8080", nil)
  }
