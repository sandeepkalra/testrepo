package main

import (
	"fmt"
	"sync"
	"time"
)

////// STATES ///////

type fsmState int

const (
	fsmStart         fsmState = 0 // "start"
	fsmInit          fsmState = 1 // "init"
)

func (s fsmState) String() string {
	switch s {
	case fsmStart:
		return "startState"
	case fsmInit:
		return "init"
	}
	return fmt.Sprintf("state: %d", int(s))
}

////// EVENTS ///////

type fsmEvent int

const (
	fsmEvtEntry fsmEvent = 0
	fsmFault    fsmEvent = 1
	fsmAbort    fsmEvent = 2
)

func (e fsmEvent) String() string {
	switch e {
	case fsmEvtEntry:
		return "entry"
	case fsmFault:
		return "error/fault"
	case fsmAbort:
		return "abort"
	}
	return fmt.Sprintf("event: %d", int(e))
}

////// FSM ////////
type FSM struct {
	fsmInternalEvents   chan fsmEvent
	fsmExternalEvents   chan fsmEvent
	fsmMutex            sync.Mutex
	fsmCurrentState     fsmState
	fsmCurrentEvent     fsmEvent
	fsmConfig           map[string]string
	fsmStateTransitions map[fsmState]map[fsmEvent]func(*FSM) (fsmState, error)
}

func (fsm *FSM) initFSM() *FSM {
	if fsm != nil {
		// this is to make it singleton
		return fsm
	}
	return &FSM{
		fsmInternalEvents: make(chan fsmEvent),
		fsmExternalEvents: make(chan fsmEvent),
		fsmCurrentEvent:   fsmEvtEntry,
		fsmCurrentState:   fsmStart,
		fsmStateTransitions: map[fsmState]map[fsmEvent]func(*FSM) (fsmState, error){
			/* STATE // EVENTS: ENTRY,                ABORT,               FAULT */
			fsmStart: {fsmEvtEntry: StartEntry, fsmAbort: StartAbort, fsmFault: StartAbort},
			fsmInit:  {fsmEvtEntry: InitEntry, fsmAbort: InitAbort, fsmFault: InitAbort},
		},
		fsmConfig: map[string]string{},
	}
}

func StartEntry(f *FSM) (fsmState, error) {
	fmt.Println("startEntryHandler")
	time.Sleep(10 * time.Second)
	f.Transition(fsmInit, fsmEvtEntry)
	return fsmStart, nil
}
func StartAbort(f *FSM) (fsmState, error) {
	fmt.Println("startAbortHandler")
	time.Sleep(10 * time.Second)
	return fsmStart, nil
}
func InitEntry(f *FSM) (fsmState, error) {
	fmt.Println("initEntryHandler")
	time.Sleep(10 * time.Second)
	f.Transition(fsmStart, fsmEvtEntry)
	return fsmStart, nil
}
func InitAbort(f *FSM) (fsmState, error) {
	fmt.Println("initAbortHandler")
	time.Sleep(10 * time.Second)
	return fsmStart, nil
}

func (fsm *FSM) Transition(next fsmState, event fsmEvent) error {
	if fsm == nil {
		return fmt.Errorf("fsm is not initialized yet")
	}
	var oldEvent fsmEvent
	var oldState fsmState

	fsm.fsmMutex.Lock()
	oldState = fsm.fsmCurrentState
	oldEvent = fsm.fsmCurrentEvent
	fsm.fsmCurrentEvent = event
	fsm.fsmCurrentState = next
	fsm.fsmMutex.Unlock()

	fmt.Printf("State transitioning from:[%v][%v] to:[%v][%v]", oldState, oldEvent, next, event)
	fsm.fsmInternalEvents <- event
	time.Sleep(3 * time.Second)
	return nil
}

func (fsm *FSM) RunThread() {
	if fsm == nil {
		fsm = fsm.initFSM()
		go func() {
			fsm.fsmInternalEvents <- fsmEvtEntry
			time.Sleep(3 * time.Second)
		}()
	}
	for {
		select {
		case e := <-fsm.fsmInternalEvents:

			fmt.Printf("event from internal world:%v.\n", e)
			fsm.fsmMutex.Lock()
			next := fsm.fsmCurrentState
			fsm.fsmMutex.Unlock()

			f, ok := fsm.fsmStateTransitions[next][e]
			if ok {
				go f(fsm)
			} else {
				fmt.Println("fail to find FSM transition for state:", next)
				time.Sleep(5 * time.Second)
			}
		case e := <-fsm.fsmExternalEvents:
			fmt.Printf("event from external world:%v.\n", e)
			fsm.fsmMutex.Lock()
			next := fsm.fsmCurrentState
			fsm.fsmMutex.Unlock()

			f, ok := fsm.fsmStateTransitions[next][e]
			if ok {
				go f(fsm)
			} else {
				fmt.Println("fail to find FSM transition for state:", next)
				time.Sleep(5 * time.Second)
			}
		case <-time.After(time.Minute * 1):
			// no process needed, this avoids panic
		} // select
	} //for
}

var gFSM *FSM

func main() {
	c := make(chan bool)
	go gFSM.RunThread()
	fmt.Println("fsm demo")
	done := <-c
	fmt.Printf("done chan got %v \n", done)
}
