package main

import (
	"fmt"
	//"go/printer"
	"io"
	//"os"
	//"strconv"
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

	var currentTimer int64 = 0
	var currentIndex int
	var nextIndex int
	var currentProcessID string
	var waitTime int64
	var startTime int64

	//new processes list used for burst time changes
	processBurstTimes := make([]Process, len(processes))
	copy(processBurstTimes, processes)

	waitTimes := make([]int64, len(processes))
	turnAroundTimes := make([]float64, len(processes))
	completionTimes := make([]float64, len(processes))

	currentIndex = findNextShortProcessIndex(0, processBurstTimes)

	currentProcessID = processes[currentIndex].ProcessID
	startTime = 0

	for {

		nextIndex = findNextShortProcessIndex(currentTimer, processBurstTimes)

		if nextIndex == -1 || nextIndex == len(processes) {

			gantt = append(gantt, TimeSlice{
				PID:   processes[currentIndex].ProcessID,
				Start: startTime,
				Stop:  currentTimer,
			})

			break
		}

		//if new process takes priority
		if processes[nextIndex].ProcessID != currentProcessID {

			//preempt current process

			//Old process' timeslice is added to the gantt table
			gantt = append(gantt, TimeSlice{
				PID:   processes[currentIndex].ProcessID,
				Start: startTime,
				Stop:  currentTimer,
			})

			//if processes is not complete
			if processBurstTimes[currentIndex].BurstDuration != 0 {
				//Time spent in the CPU does not add to wait time
				waitTimes[currentIndex] += (currentTimer - startTime) * -1
			}

			//load new process
			currentIndex = nextIndex
			startTime = currentTimer

			currentProcessID = processBurstTimes[currentIndex].ProcessID

			//new process' wait time is calculated
			waitTime = currentTimer - processes[currentIndex].ArrivalTime
			waitTimes[currentIndex] += waitTime
			totalWait += float64(waitTimes[currentIndex])

			//partial turnaround and completion times calculated
			turnAroundTimes[currentIndex] += float64(waitTimes[currentIndex])
			completionTimes[currentIndex] += float64(waitTimes[currentIndex])
		}

		//Decrementing amount processes has left if same process continues
		processBurstTimes[currentIndex].BurstDuration -= 1

		currentTimer++

	}

	for i := 0; i < len(processes); i++ {

		//complete turnaround times calculated for all processes
		turnAroundTimes[i] += float64(processes[i].BurstDuration)
		totalTurnaround += turnAroundTimes[i]

		//complete completion times calculated for all processes
		completionTimes[i] += float64(processes[i].BurstDuration) + float64(processes[i].ArrivalTime)

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
	aveThroughput := count / (float64(currentTimer))

	//Final outputs
	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)

}

func findNextShortProcessIndex(currentTime int64, processes []Process) int {

	var processWaiting bool
	var processIsShortest bool

	var shortestProcessIndex int = -1

	for i := 0; i < len(processes); i++ {

		//skip processes with no execution needed
		if processes[i].BurstDuration <= 0 {
			continue
		}

		//initiate to first value with CPU time needed
		if shortestProcessIndex == -1 {
			shortestProcessIndex = i
			continue
		}

		processWaiting = processes[i].ArrivalTime <= int64(currentTime)

		processIsShortest = processes[i].BurstDuration < processes[shortestProcessIndex].BurstDuration

		if processWaiting && processIsShortest {
			shortestProcessIndex = i
		}
	}

	return shortestProcessIndex

}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {}

func RRSchedule(w io.Writer, title string, processes []Process) {}

//endregion
