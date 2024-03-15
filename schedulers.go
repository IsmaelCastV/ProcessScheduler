package main

import (
	"fmt"
	//"go/printer"
	"io"
	"os"
)

type (
	Process struct {
		ProcessID     string
		ArrivalTime   int64
		BurstDuration int64
		Priority      int64
	}
	TimeSlice struct {
		PID   string
		Start int64
		Stop  int64
	}
)

//region Schedulers

// FCFSSchedule outputs a schedule of processes in a GANTT chart and a table of timing given:
// • an output writer
// • a title for the chart
// • a slice of processes
func FCFSSchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		waitingTime     int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)
	for i := range processes {

		//if process i is greater arrives after the first process
		if processes[i].ArrivalTime > 0 {
			//waitTime = time process i entered CPU - time processes arrived
			waitingTime = serviceTime - processes[i].ArrivalTime
		}

		//totalwait = converted wait time
		totalWait += float64(waitingTime)

		//process i starts at waitTime + process arrival time
		start := waitingTime + processes[i].ArrivalTime

		//process turnaround = waitTime + time spent on CPU
		turnaround := processes[i].BurstDuration + waitingTime
		totalTurnaround += float64(turnaround)

		//process comletion = CPU time + arrival time + wait time
		completion := processes[i].BurstDuration + processes[i].ArrivalTime + waitingTime
		lastCompletion = float64(completion)

		//scheduler holds process on first come first serve basis
		schedule[i] = []string{
			fmt.Sprint(processes[i].ProcessID),
			fmt.Sprint(processes[i].Priority),
			fmt.Sprint(processes[i].BurstDuration),
			fmt.Sprint(processes[i].ArrivalTime),
			fmt.Sprint(waitingTime),
			fmt.Sprint(turnaround),
			fmt.Sprint(completion),
		}

		//increment serviceTime by process i CPU time
		serviceTime += processes[i].BurstDuration

		//process added to gantt chart
		gantt = append(gantt, TimeSlice{
			PID:   processes[i].ProcessID,
			Start: start,
			Stop:  serviceTime,
		})
	}

	//number of processes
	count := float64(len(processes))

	//average wait time
	aveWait := totalWait / count

	//average ( CPU time + waitTime) / count
	aveTurnaround := totalTurnaround / count

	//average process / time
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func SJFSchedule(w io.Writer, title string, processes []Process) {
	/*var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		waitingTime     int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)*/

	fmt.Println("*****Before Sorting******")

	for i := range processes {

		fmt.Println("Processes ID: ", processes[i].ProcessID, "| Duration: ", processes[i].BurstDuration)
	}

	sortByBurstDesc(processes)

	fmt.Println("        ")

	fmt.Println("*****After Sorting******")

	for i := range processes {

		fmt.Println("Processes ID: ", processes[i].ProcessID, "| Duration: ", processes[i].BurstDuration)
	}
}

func sortByBurstDesc(processes []Process) {

	var temp Process

	for i := 0; i < len(processes)-1; i++ {

		if i > 100 {
			os.Exit(100)

		}

		for j := i + 1; j < len(processes); j++ {

			if j > 200 {
				os.Exit(200)
			}

			if processes[j].BurstDuration < processes[i].BurstDuration {

				temp = processes[i]

				processes[i] = processes[j]

				processes[j] = temp

				fmt.Println("*****Changes******")

				fmt.Println("i: ", i, "j: ", j)

				for i := range processes {

					fmt.Println("Processes ID: ", processes[i].ProcessID, "| Duration: ", processes[i].BurstDuration)
				}

			}
		}

	}

}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {}

func RRSchedule(w io.Writer, title string, processes []Process) {}

//endregion
