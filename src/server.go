package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
)

/*
 * @Author: Rouzip
 * @Date: 2020-12-11 23:22:32
 * @LastEditTime: 2020-12-12 21:32:50
 * @LastEditors: Rouzip
 * @Description: My blog webhook server
 */

// GitRepo the detail of the git repo
type GitRepo struct {
	Respo GitName
}

// GitName url and the name of the repo
type GitName struct {
	CloneURL string `json:"clone_url"`
	Name     string `json:"name"`
}

func readFile(name string) ([]byte, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
	}
	return data, err
}

func parseJSONStr(data []byte) map[string]interface{} {
	dataMap := make(map[string]interface{})
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.Decode(&dataMap)
	return dataMap
}

func getEnv(jsonStr map[string]interface{}, key string) string {
	if str, ok := jsonStr[key]; ok {
		return str.(string)
	}
	panic("invalid value")

}

func loadConfPath() *string {
	return flag.String("f", "env.conf", "the config of the webhook")
}

func getJSON() map[string]interface{} {
	path := loadConfPath()
	if path == nil {
		panic("path is invalid!")
	}
	if data, err := readFile(*path); err == nil {
		return parseJSONStr(data)
	}
	return nil
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
	envJSON := getJSON()
	key := getEnv(envJSON, "KEY")
	blogIndex := getEnv(envJSON, "PATH")
	PORT := getEnv(envJSON, "PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		sign := r.Header.Get("x-hub-signature")
		if checkSum(key, sign, r.Body) {
			decoder := json.NewDecoder(r.Body)
			gitDetail := &GitRepo{}
			decoder.Decode(gitDetail)

			cmd := exec.Command("/bin/sh", "-c", "cd /tmp; git clone "+gitDetail.Respo.CloneURL+";")
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "mv /tmp/"+gitDetail.Respo.Name+"/md/* "+blogIndex+"/content/post")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "mv /tmp/"+gitDetail.Respo.Name+"/img/* "+blogIndex+"/static")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			cmd = exec.Command("/bin/sh", "-c", "cd "+blogIndex+";rm -rf public;hugo;")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
	})
	http.ListenAndServe("0.0.0.0:"+PORT, mux)
}
