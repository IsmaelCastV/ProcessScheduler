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

		//if process i is not the first process
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
	var (
		totalWait       float64
		totalTurnaround float64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)

	// fmt.Println("*****Before Sorting******")

	// for i := range processes {

	// 	fmt.Println("Processes ID: ", processes[i].ProcessID, "| Duration: ", processes[i].BurstDuration)
	// }

	//REFACTOR START

	var currentTimer int64 = 0
	var currentIndex int
	var nextIndex int
	var currentProcessID string
	var currentProcessBursts int64
	var waitTime int64
	var startTime int64

	processBurstTimes := make([]Process, len(processes))
	copy(processBurstTimes, processes)

	waitTimes := make([]int64, len(processes))
	turnAroundTimes := make([]float64, len(processes))
	completionTimes := make([]float64, len(processes))

	//TEST:
	// processes[1].BurstDuration = 0

	// fmt.Println("Hard Code: ", processes[1].ProcessID, " Duration:  ", processes[1].BurstDuration)
	var loopControl = 50
	//END OF TEST

	currentIndex = findNextShortProcessIndex(0, processBurstTimes)

	currentProcessID = processes[currentIndex].ProcessID
	currentProcessBursts = processes[currentIndex].BurstDuration
	startTime = 0

	for {

		//DEBUG SECTION
		if loopControl == 0 {
			os.Exit(loopControl)
		}
		//END DEBUG SECTION

		fmt.Println("#####Current Time: ", currentTimer, "#####")
		fmt.Println("\tCurrent Process: ", currentProcessID)
		fmt.Println("\tDuration: ", currentProcessBursts)

		nextIndex = findNextShortProcessIndex(currentTimer, processBurstTimes)

		fmt.Println("nextIndex: ", nextIndex)

		if nextIndex == -1 || nextIndex == len(processes) {

			gantt = append(gantt, TimeSlice{
				PID:   processes[currentIndex].ProcessID,
				Start: startTime,
				Stop:  currentTimer,
			})

			break
		}

		fmt.Println("")

		if processes[nextIndex].ProcessID != currentProcessID {

			//preempt current process

			fmt.Println("------->*****Preemptive Action*****")
			fmt.Println("\t#\tOld Process: ", currentProcessID)
			fmt.Println("\t#\tNew Duration: ", currentProcessBursts)

			//process added to gantt chart
			gantt = append(gantt, TimeSlice{
				PID:   processes[currentIndex].ProcessID,
				Start: startTime,
				Stop:  currentTimer,
			})

			//load new processes

			currentIndex = nextIndex
			startTime = currentTimer

			currentProcessID = processBurstTimes[currentIndex].ProcessID
			currentProcessBursts = processBurstTimes[currentIndex].BurstDuration

			waitTime = currentTimer - processes[currentIndex].ArrivalTime
			waitTimes[currentIndex] += waitTime
			totalWait += float64(waitTime)

			turnAroundTimes[currentIndex] += float64(waitTime)
			completionTimes[currentIndex] += float64(waitTime)

			fmt.Println("\t#New Process: ", currentProcessID)
			fmt.Println("\t#\tDuration: ", currentProcessBursts)
			fmt.Println("\t#\tCurrent Wait Time: ", waitTimes[currentIndex], "@index: ", currentIndex)
			fmt.Println("")

		}

		processBurstTimes[currentIndex].BurstDuration -= 1
		currentProcessBursts = processBurstTimes[currentIndex].BurstDuration

		currentTimer++

		loopControl--

	}

	for i := 0; i < len(processes); i++ {

		turnAroundTimes[i] += float64(processes[i].BurstDuration)
		totalTurnaround += turnAroundTimes[i]

		completionTimes[i] += float64(processes[i].BurstDuration) + float64(processes[i].ArrivalTime)

		fmt.Println("Process[ ", i, " ]: ", processes[i].BurstDuration)

		//scheduler holds process on first come first serve basis
		schedule[i] = []string{
			fmt.Sprint(processes[i].ProcessID),
			fmt.Sprint(processes[i].Priority),
			fmt.Sprint(processes[i].BurstDuration),
			fmt.Sprint(processes[i].ArrivalTime),
			fmt.Sprint(waitTimes[i]),
			fmt.Sprint(turnAroundTimes[i]),
			fmt.Sprint(completionTimes[i]),
		}
	}

	//number of processes
	count := float64(len(processes))

	//average wait time
	aveWait := totalWait / count

	//average ( CPU time + waitTime) / count
	aveTurnaround := totalTurnaround / count

	//average process / time
	aveThroughput := count / (float64(currentTimer) - 1)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)

}

func findNextShortProcessIndex(currentTime int64, processes []Process) int {

	if len(processes) == 0 {

		return -1
	}

	var processNeedsCPU bool
	var processWaiting bool
	var processIsShortest bool

	var shortestProcessIndex int = -1

	//fmt.Println("Largest Value: ", shortestBurst)

	fmt.Println("\t******Finding Next Process******")

	for i := 0; i < len(processes); i++ {

		// if processes[i].BurstDuration == 0 {
		// 	processes = removeItem(processes, i)
		// }

		if processes[i].BurstDuration == 0 {

			continue
		}

		if shortestProcessIndex == -1 {

			shortestProcessIndex = i
			continue
		}

		processNeedsCPU = processes[i].BurstDuration > 0

		processWaiting = processes[i].ArrivalTime <= int64(currentTime)

		processIsShortest = processes[i].BurstDuration < processes[shortestProcessIndex].BurstDuration

		// fmt.Println("\tProcess: ", processes[i].ProcessID, " | Duration: ", processes[i].BurstDuration)
		// fmt.Println("\tProcess Needs CPU: ", processNeedsCPU, " | Process In Queue: ", processWaiting, " | Process is Shortest: ", processIsShortest)
		// fmt.Println("\t*****")

		if processNeedsCPU && processWaiting && processIsShortest {

			shortestProcessIndex = i
			fmt.Println("\tUpdate: Shortest Job: ", processes[shortestProcessIndex].ProcessID, " Duration: ", processes[shortestProcessIndex].BurstDuration)

		}
	}

	//fmt.Println("Here's ya problem: ", shortestProcessIndex)

	//fmt.Println("\tATT: Active Process: ", processes[shortestProcessIndex].ProcessID, " | Index: ", shortestProcessIndex)
	return shortestProcessIndex

}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {}

func RRSchedule(w io.Writer, title string, processes []Process) {}

//endregion
