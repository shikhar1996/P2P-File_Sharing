package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type File_s struct {
	Name string
	Hash string
	Size float64
	path string
	ip   string
}

var files []File_s
var ips []string
var LIST_URL = "http://ayushpateria.com/ShareIIT/list.php"

//Define that the binairy data of the file will be sent 1024 bytes at a time
const BUFFERSIZE = 81920

func fetchIPS() {
	ips = nil
	response, err := http.Get(LIST_URL)
	if err != nil {
		fmt.Println(err)
	} else {
		defer response.Body.Close()
		r, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(r, &ips)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func updateList(ip string) {
	var lFiles []File_s
	connection, err := net.Dial("tcp", ip+":3333")
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Fprintf(connection, "1\n")
	message, _ := bufio.NewReader(connection).ReadString('\n')

	json.Unmarshal([]byte(message), &lFiles)
	for i, _ := range lFiles {
		lFiles[i].ip = ip
	}
	files = append(files, lFiles...)
}

func createList() {
	files = nil
	fetchIPS()
	var wg sync.WaitGroup
	for _, ip := range ips {
		// Increment the WaitGroup counter.
		wg.Add(1)
		go func(ip string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			updateList(ip)
		}(ip)
	}
	// Wait for all Lists fetches to complete.
	wg.Wait()
}

func main() {

	fmt.Println("Welcome to ShareIIT! Your intra college file sharing hub.")

	fmt.Println("1. List all available files.")
	fmt.Println("2. Download a file,")
	fmt.Println("3. Search a file.")
	fmt.Println("0. Exit.")
	for {

		fmt.Print(">> ")

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		option, _ := reader.ReadString('\n')
		if option[:1] == "1" {

			createList()

			for i, value := range files {
				fmt.Print((i + 1))
				fmt.Print(". " + value.Name + "		")
				fmt.Print(value.Size)
				fmt.Println(" kb")
			}

		} else if option[:1] == "2" {

			if len(files) == 0 {
				createList()
			}
			fmt.Println("Enter the ID of the file from the list : ")
			var id int
			fmt.Scanf("%d", &id)
			if id > len(files) {
				fmt.Println("Please enter a valid ID.")
			} else {
				receivefile(id - 1)
			}
		} else if option[:1] == "3" {
		
			createList()
			
			fmt.Print("Enter file name : ")
			reader := bufio.NewReader(os.Stdin)
			filename, _ := reader.ReadString('\n')
			flag := 0
			//fmt.Print(filename)
			for i, value := range files {
				//fmt.Println(value.Name)
				if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
					fmt.Print((i + 1))
					fmt.Print(". " + value.Name + "		")
					fmt.Print(value.Size)
					fmt.Println(" kb")
					flag = 1
				}
			}
			if flag == 0 {
				fmt.Print("No items match your search.\n")
			}
		} else if option[:1] == "0" {
			break
		}
	}

}

func receivefile(i int) {
	file := files[i]
	fmt.Println("Downloading " + file.Name + ", this may take a while.")
	connection, err := net.Dial("tcp", file.ip+":3333")
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	hash := files[i].Hash
	fmt.Fprintf(connection, "2 "+hash+"\n")
	//Create buffer to read in the name and size of the file
	bufferFileName := make([]byte, 128)
	bufferFileSize := make([]byte, 20)
	//Get the filesize
	connection.Read(bufferFileSize)
	//Strip the ':' from the received size, convert it to a int64
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 20, 64)
	//Get the filename
	connection.Read(bufferFileName)
	//Strip the ':' once again but from the received file name now
	fileName := strings.Trim(string(bufferFileName), ":")
	//Create a new file to write in
	newFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	//Create a variable to store in the total amount of data that we received already
	var receivedBytes int64
	//Start writing in the file
	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			//Empty the remaining bytes that we don't need from the network buffer
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			//We are done writing the file, break out of the loop
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		//Increment the counter
		receivedBytes += BUFFERSIZE
	}
	fmt.Println("Received file completely!")
}
