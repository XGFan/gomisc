package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/mozillazg/go-cos"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var debug bool
var prefix string
var random bool
var stdout bool
var board bool

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func main() {
	flag.BoolVar(&debug, "v", false, "enable debug log")
	flag.StringVar(&prefix, "p", "temp/", "prefix")
	flag.BoolVar(&random, "r", false, "random name")
	flag.BoolVar(&stdout, "o", false, "write url to output")
	flag.BoolVar(&board, "c", true, "write url to clipBoard")
	flag.Parse()
	var urls []string = make([]string, 0)
	for _, fileName := range flag.Args() {
		file, err := getFile(fileName)
		if err != nil {
			log.Fatal(err)
		}
		url, err := PutFileToCos(file)
		if err != nil {
			log.Fatal(err)
		}
		urls = append(urls, url)
	}
	allUrl := strings.Join(urls, "\n")
	if stdout {
		fmt.Println(allUrl)
	}
	if board {
		_ = clipboard.WriteAll(allUrl)
	}

}

func getFile(fileName string) (file string, e error) {
	if debug {
		log.Printf("get file %s", fileName)
	}
	dir, _ := os.Getwd()                 //get current dir
	join := filepath.Join(dir, fileName) //get relative path
	if fileExists(join) {
		file = join
	} else {
		if fileExists(fileName) {
			file = fileName
			if debug {
				log.Println("absolute yes")
			}
		} else {
			e = errors.New("file not found")
		}
	}
	return
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadCosConfig() map[string]string {
	home := os.Getenv("HOME")
	file := home + "/.cos"
	bytes, _ := ioutil.ReadFile(file)
	s := string(bytes)
	split := strings.Split(s, "\n")
	m := make(map[string]string)
	for _, v := range split {
		if v == "" {
			continue
		}
		i := strings.IndexRune(v, '=')
		runes := []rune(v)
		name := strings.TrimSpace(string(runes[0:i]))
		value := strings.TrimSpace(string(runes[i+1:]))
		m[name] = value
	}
	return m
}

func PutFileToCos(fileName string) (url string, e error) {
	config := LoadCosConfig()
	if debug {
		log.Printf("config: %v", config)
		log.Printf("file: %s", fileName)
	}
	b, _ := cos.NewBaseURL(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config["bucket"], config["region"]))
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config["secretId"],
			SecretKey: config["secretKey"],
		},
	})
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		e = err
		return
	}
	var key string
	prefix = strings.ReplaceAll(prefix, "{yyyy}", strconv.Itoa(time.Now().Year()))
	prefix = strings.ReplaceAll(prefix, "{mm}", fmt.Sprintf("%02d", int(time.Now().Month())))
	prefix = strings.ReplaceAll(prefix, "{dd}", fmt.Sprintf("%02d", time.Now().Day()))
	if random {
		split := strings.Split(stat.Name(), ".")
		extension := split[len(split)-1]
		if extension != "" {
			key = fmt.Sprintf("%s%s.%s", prefix, String(8), extension)
		} else {
			key = fmt.Sprintf("%s%s", prefix, String(8))
		}
	} else {
		key = fmt.Sprintf("%s%s", prefix, stat.Name())
	}
	resp, err := c.Object.Put(context.Background(), key, f, nil)
	if err != nil {
		panic(err)
	}
	if config["domain"] != "" {
		url = config["domain"] + key
	} else {
		url = resp.Header.Get("Location")
	}

	return
}

func SHA1(value []byte) string {
	hash := sha1.New()
	hash.Write(value)
	return hex.EncodeToString(hash.Sum(nil))
}
func HMACSHA1(input, key string) string {
	keyForSign := []byte(key)
	h := hmac.New(sha1.New, keyForSign)
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
