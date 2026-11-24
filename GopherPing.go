// Concurrent Network & TCP Port Scanner

package main

import (
	// "fmt"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
	"os"
)

type ScanResult struct{
	Target string `json:"target"`
	Status string `json:"status"`
	Duration string `json:"duration"`
}

//Allows the function to be able to return the struct values
func scanPort(TString string) ScanResult{
	var instance ScanResult	
	Time:=time.Now()

	//attaching a strict stopwatch to your connection attempt to prevent the Go program from freezing
	//Accepts 3 Arguments, 
	// 1.The Protocol ("tcp"). This tells Go to use the standard internet rules to verify connection.
	// 2. The Target & Port string("google.com:443"): The destination address combined with the specific port number you want to test, separated by a colon.
	// The Maximum Wait Duration (2 * time.Second): Your stopwatch limit, if does not answer within 2 seconds, abort the mission."
	//DialTimeout means "place a network call, but hang up automatically if nobody answers before the timer runs out."
	t_conn, err :=net.DialTimeout("tcp", TString, 2*time.Second)
	delay:=time.Since(Time)

	if err!=nil{
		fmt.Println(err)
		instance.Target=TString
		instance.Status="BLOCKED"
		instance.Duration=delay.String()
		
		fmt.Printf("\nConnection to %s, %s after %s", instance.Target, instance.Status, instance.Duration)
		return instance
		// fmt.Printf("\n[BLOCKED/TIMEOUT] %s (took %v sec)\n", t_conn, delay)
	}else{
		t_conn.Close()
		instance.Target=TString
		instance.Status="OPEN"
		instance.Duration=delay.String()

		return instance
	}
}



func main(){
	end_points:=[]string{"google.com:443", "google.com:22", "github.com:443", "amazon.com:80"}

	//Creates a buffered channel(buff-FIFO queue) of type  ScanResult
	//Goroutines cant return values to the main function, so channels act as memory pipes that allow workers to drop of their ScanResult structs
	buff:=make(chan ScanResult, len(end_points))

	//Tells the main that it should launch all the gorountines and wait for them to finish before terminating
	//A lock to prevent the main from terminating
	var wg sync.WaitGroup
	
	//Loops throught as much time as they are arrays stored
	for _,server:=range end_points{
		//Background worker, makes the digital counter(wg) increment by one
		//Creaeted before each goroutine runs, so the main knows how many blocks to wait(wg.Wait) for before terminating
		wg.Add(1)
		
		// Creates a goroutine(a background process that can run simultaneously)
		//(t string) part creates a temporary, isolated variable slot inside that background worker
		go func(t string){
			defer wg.Done()
			//This guarantees that no matter what happens during the network scan, the counter will always click down when the function exits.
			// Written before the scanport becuase of runtime panic, thats why defer is used
			//function call
			CapturedStruct:=scanPort(t)
			buff<-CapturedStruct
		}(server)//Argument that is being passed as t string to the fake function(Nameless function)
	}

	wg.Wait()
	//Channels have no close method they are closed using the close function because they are built in
	close(buff)

	var Data []ScanResult
	//Channels can only give a single variable, they have no index
	for n:=range buff{
		Data=append(Data, n)
	}

	//Takes three arguments the value to encode (Data), a line prefix (""), and an indentation string ("  ").
	//what to put at the start of every line ("" for nothing) and how many spaces to indent each level ("  ").
	Formatted, err:=json.MarshalIndent(Data, "", " ")

	if err!=nil{
		fmt.Println(err)
	}else{
		//Creaetes a file, puts data into that file and gives permisiions for that file
		err=os.WriteFile("audit_report.json", Formatted, 0644)
		if err!=nil{
			fmt.Println(err)
		}
	}
}