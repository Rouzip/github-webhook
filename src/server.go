package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/tidwall/gjson"
)

/*
 * @Author: Rouzip
 * @Date: 2020-12-11 23:22:32
 * @LastEditTime: 2020-12-15 01:03:36
 * @LastEditors: Rouzip
 * @Description: My blog webhook server
 */

func readFile(name string) ([]byte, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
	}
	return data, err
}

func loadConfPath() *string {
	return flag.String("c", "env.conf", "the config of the webhook")
}

func checkSum(key, sign string, data io.ReadCloser) bool {
	h := hmac.New(sha1.New, []byte(key))
	body, err := ioutil.ReadAll(data)
	if err != nil {
		fmt.Println(err)
	}
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil)) == sign[5:]
}

func main() {
	path := loadConfPath()
	confBytes, err := readFile(*path)
	if err != nil {
		fmt.Println(err)
	}
	confStr := string(confBytes)
	key := gjson.Get(confStr, "KEY")
	blogIndex := gjson.Get(confStr, "PATH")
	port := gjson.Get(confStr, "PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		sign := r.Header.Get("x-hub-signature")
		if checkSum(key.String(), sign, r.Body) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Println(err)
			}
			bodyStr := string(body)
			gitURL := gjson.Get(bodyStr, "repository.clone_url")
			name := gjson.Get(bodyStr, "repository.name")

			cmd := exec.Command("/bin/sh", "-c", "cd /tmp; git clone "+gitURL.String()+";")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "mv /tmp/"+name.String()+"/md/* "+blogIndex.String()+"/content/post")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "mv /tmp/"+name.String()+"/img/* "+blogIndex.String()+"/static")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "cd "+blogIndex.String()+";rm -rf public;hugo;")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
	})
	http.ListenAndServe("0.0.0.0:"+port.String(), mux)
}
