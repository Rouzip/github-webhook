package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
)

/*
 * @Author: Rouzip
 * @Date: 2020-12-11 23:22:32
 * @LastEditTime: 2020-12-15 14:17:49
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

func checkSum(key, sign string, data []byte) bool {
	h := hmac.New(sha1.New, []byte(key))
	h.Write(data)
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
	blogIndex := strings.TrimSpace(gjson.Get(confStr, "PATH").String())
	port := gjson.Get(confStr, "PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		sign := r.Header.Get("x-hub-signature")
		bodyBytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		if checkSum(key.String(), sign, bodyBytes) {
			bodyStr := string(bodyBytes)
			gitURL := strings.TrimSpace(gjson.Get(bodyStr, "repository.clone_url").String())
			name := strings.TrimSpace(gjson.Get(bodyStr, "repository.name").String())

			output, err := exec.Command("/bin/sh", "-c", "cd /tmp; git clone "+gitURL+";").CombinedOutput()
			if err != nil {
				fmt.Println("Error when running command.  Output:")
				fmt.Println(string(output))
				fmt.Printf("Got command status: %s\n", err.Error())
			}
			output, err = exec.Command("/bin/sh", "-c", "mv /tmp/"+name+"/md/* "+blogIndex+"/content/post").CombinedOutput()
			if err != nil {
				fmt.Println("Error when running command.  Output:")
				fmt.Println(string(output))
				fmt.Printf("Got command status: %s\n", err.Error())
			}
			output, err = exec.Command("/bin/sh", "-c", "mv /tmp/"+name+"/img/* "+blogIndex+"/static").CombinedOutput()
			if err != nil {
				fmt.Println("Error when running command.  Output:")
				fmt.Println(string(output))
				fmt.Printf("Got command status: %s\n", err.Error())
			}
			output, err = exec.Command("/bin/sh", "-c", "cd "+blogIndex+";rm -rf public;hugo;").CombinedOutput()
			if err != nil {
				fmt.Println("Error when running command.  Output:")
				fmt.Println(string(output))
				fmt.Printf("Got command status: %s\n", err.Error())
			}
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
	})
	http.ListenAndServe("0.0.0.0:"+port.String(), mux)
}
