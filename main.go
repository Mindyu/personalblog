package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"
)

/*
go 调用mongodb
http 服务器
安装
go get -u gopkg.in/mgo.v2
*/

var mgosess *mgo.Session

type Person struct {
	Name string
	Age  int
}

type Blog struct {
	Title   string
	FileDir string
}

//响应消息结构
type Resp struct {
	Errno  string
	Errmsg string
	Data   interface{}
}

func pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

//统一响应消息
func RespJson(resp *Resp, w http.ResponseWriter) {
	resp.Errno = "0"
	if resp.Errmsg != "ok" {
		resp.Errno = "100" // 应该有错误编码
	}
	//转换为json
	data, _ := json.Marshal(resp)
	w.Write(data)
}

func publish(w http.ResponseWriter, r *http.Request) {
	//0. 响应消息
	var resp Resp
	resp.Errmsg = "ok"
	defer RespJson(&resp, w)
	//1. 获取前端消息
	title := r.FormValue("title")
	content := r.FormValue("content")
	//正文特别多，怎么办？生成文件
	filehash := sha256.Sum256([]byte(title))
	filename := fmt.Sprintf("%x", filehash)
	fmt.Println(filename)
	f, err := os.Create("static/blogfile/" + filename)
	if err != nil {
		fmt.Println("failed to create file ", err)
		resp.Errmsg = err.Error()
		return
	}
	defer f.Close()
	f.WriteString(content) //写入文件
	//2. 保存到数据库
	blog := Blog{title, "blogfile/" + filename}
	table := mgosess.DB("yekai").C("blogs")
	err = table.Insert(blog)
	if err != nil {
		fmt.Println("failed to insert mongo", err)
		resp.Errmsg = err.Error()
		return
	}
}

func lists(w http.ResponseWriter, r *http.Request) {
	//1. 组织响应消息
	var resp Resp
	resp.Errmsg = "ok"
	defer RespJson(&resp, w)
	//2. 查询数据库
	table := mgosess.DB("yekai").C("blogs")
	var blogs []Blog
	err := table.Find(nil).All(&blogs)
	if err != nil {
		fmt.Println("failed to query mongo", err)
		resp.Errmsg = err.Error()
		return
	}
	//3. 响应消息赋值处理
	resp.Data = blogs
}

func main() {
	fmt.Println("hello every one!")
	//连接到mongodb
	sess, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println("failed to connect to mongo", err)
		return
	}
	defer sess.Close() // 本函数返回时自动执行
	mgosess = sess
	//	//1. 增加mongodb 一条记录
	//	table := sess.DB("yekai").C("langzi") // 获得集合
	//	table.Insert(bson.M{"name": "yekai", "age": 36})
	//	//2. 查询
	//	var persons []Person
	//	table.Find(nil).All(&persons)
	//	fmt.Println(persons)

	http.HandleFunc("/ping", pong)                        //测试用
	http.Handle("/", http.FileServer(http.Dir("static"))) //提供静态文件服务的根目录
	http.HandleFunc("/publish", publish)                  //发表博客
	http.HandleFunc("/lists", lists)                      //发表博客
	http.ListenAndServe(":8086", nil)                     //启动http服务器
}
