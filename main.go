package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var mgosess *mgo.Session

//响应消息结构
type Resp struct {
	Errno  string      //错误码
	Errmsg string      //错误消息
	Data   interface{} //携带正文
}
type BlogInfo struct {
	Title   string
	FileDir string
}

//定义通用响应消息结构
func RespJson(resp *Resp, w http.ResponseWriter) {
	resp.Errno = "0"
	if resp.Errmsg != "ok" {
		resp.Errno = "100"
	}
	data, _ := json.Marshal(resp)
	w.Write(data)
}

func main() {
	fmt.Println("hello everyone!")
	//mongodb 链接
	sess, err := mgo.Dial("localhost:27017")

	if err != nil {
		fmt.Println("failed to connect to mongo", err)
		return
	}
	mgosess = sess
	defer sess.Close()

	//设置路由
	http.HandleFunc("/ping", pong)
	http.Handle("/", http.FileServer(http.Dir("static"))) // 设置静态文件服务
	http.HandleFunc("/publish", publish)
	http.HandleFunc("/lists", lists)
	http.ListenAndServe(":8086", nil)
}

//服务器测试
func pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

//发表博客
func publish(w http.ResponseWriter, r *http.Request) {
	//1. 组织响应消息
	var resp Resp
	resp.Errmsg = "ok"
	defer RespJson(&resp, w)
	//2. 读取请求数据
	title := r.FormValue("title")
	content := r.FormValue("content")
	//3. 将博客正文写到文件中，使用hash算法
	filehash := sha256.Sum256([]byte(title))
	filename := fmt.Sprintf("%x", filehash)
	fmt.Println(filename)
	//fileDir := fmt.Sprintf("static/blogfile/%x", filename)
	f, err := os.Create("static/blogfile/" + filename)
	if err != nil {
		fmt.Println("failed to open file", err)
		resp.Errmsg = err.Error()
		return
	}
	defer f.Close()
	f.WriteString(content)
	//4. 将博客信息插入到数据库当中
	blog := BlogInfo{title, "blogfile/" + filename}

	table := mgosess.DB("yekai").C("blogs")
	err = table.Insert(blog)
	if err != nil {
		fmt.Println("failed to insert blog", err)
		resp.Errmsg = err.Error()
		return
	}
	//w.Write([]byte("pong\n"))
}

//查看博客列表
func lists(w http.ResponseWriter, r *http.Request) {
	//1. 组织响应消息
	var resp Resp
	resp.Errmsg = "ok"
	defer RespJson(&resp, w)

	//2. 查看数据库
	table := mgosess.DB("yekai").C("blogs")
	var blogs []BlogInfo
	err := table.Find(nil).All(&blogs)
	if err != nil {
		resp.Errmsg = err.Error()
		return
	}
	resp.Data = blogs
}
