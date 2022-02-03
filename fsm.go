package main

import (
	"fmt"
	"sync"
	"time"
)

////// STATES ///////

// FSMStateType is data-type of fsmState
type FSMStateType int

const (
	// FSMStartState is the start of this FSM
	FSMStartState FSMStateType = 0 // "start"
	// FSMInitState is a initialization state, basically for demo here,
	// and all it does is transitions back to FSMStartState
	FSMInitState FSMStateType = 1 // "init"
)

func (s FSMStateType) String() string {
	switch s {
	case FSMStartState:
		return "startState"
	case FSMInitState:
		return "init"
	}
	return fmt.Sprintf("unknown_state: %d", int(s))
}

////// EVENTS ///////

//FSMEventType is data-type of fsmEvents
type FSMEventType int

const (
	fsmEvtEntry FSMEventType = 0
	fsmFault    FSMEventType = 1
	fsmAbort    FSMEventType = 2
)

func (e FSMEventType) String() string {
	switch e {
	case fsmEvtEntry:
		return "entry"
	case fsmFault:
		return "error/fault"
	case fsmAbort:
		return "abort"
	}
	return fmt.Sprintf("unknown_event: %d", int(e))
}

////// FSM ////////

// FSM is the main data structure to hold FSM data
type FSM struct {
	fsmInternalEvents       chan FSMEventType
	fsmExternalEvents       chan FSMEventType
	fsmMutex                sync.Mutex
	fsmCurrentState         FSMStateType
	fsmCurrentEvent         FSMEventType
	fsmConfig               map[string]string
	FSMStateTypeTransitions map[FSMStateType]map[FSMEventType]func(*FSM) (FSMStateType, error)
}

func (fsm *FSM) initFSM() *FSM {
	if fsm != nil {
		// this is to make it singleton
		return fsm
	}
	return &FSM{
		fsmInternalEvents: make(chan FSMEventType),
		fsmExternalEvents: make(chan FSMEventType),
		fsmCurrentEvent:   fsmEvtEntry,
		fsmCurrentState:   FSMStartState,
		FSMStateTypeTransitions: map[FSMStateType]map[FSMEventType]func(*FSM) (FSMStateType, error){
			/* STATE // EVENTS: ENTRY,                ABORT,               FAULT */
			FSMStartState: {fsmEvtEntry: StartEntry, fsmAbort: StartAbort, fsmFault: StartAbort},
			FSMInitState:  {fsmEvtEntry: InitEntry, fsmAbort: InitAbort, fsmFault: InitAbort},
		},
		fsmConfig: map[string]string{},
	}
}

// InternalStateEvent handles internal state transtions.
// By "Internal" it means that functions/state-handlers
// direct the FSM to new states/events and they are taken
// care by this function.
func (fsm *FSM) InternalStateEvent(next FSMStateType, event FSMEventType) error {
	if fsm == nil {
		return fmt.Errorf("fsm is not initialized yet")
	}
	var oldEvent FSMEventType
	var oldState FSMStateType

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

// ExternalStateEvent is an interface function that helps external objects post
// events to the FSM.
func (fsm *FSM) ExternalStateEvent(event FSMEventType) error {
	// for External Events, we do not know what is the state, so we fetch the current state
	// and then simply process the event.
	if fsm == nil {
		return fmt.Errorf("fsm is not initialized yet, external event %v dropped", int(event))
	}

	fsm.fsmExternalEvents <- event
	time.Sleep(3 * time.Second)
	return nil
}

// RunThread is the main thread of FSM.
// it performs two main task. (a) If fsm is not created yet, it creates it oe else use it
// and (b) It has a loop that constantly monitor for events internally or external to system
// and then transition the states after processing those events.
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

			f, ok := fsm.FSMStateTypeTransitions[next][e]
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

			f, ok := fsm.FSMStateTypeTransitions[next][e]
			if ok {
				go f(fsm)
			} else {
				fmt.Println("fail to find FSM transition for state:", next)
				time.Sleep(5 * time.Second)
			}
		case <-time.After(5 * time.Minute): // no process needed,
			// this case of time.After() avoids !deadlock!
			// as there is no thread working to put the message
			// on the 2 event channels 'iff' statemachine transition
			// table is empty or has no place to proceed.
			// Also: do not put "default" case, as that gets unblocked
			// every iteration and is just CPU spin for no reason
		} // select
	} //for
}

// gFSM is global variable that points to the FSM struct.
// we need this pointer in global space so that any other external object
// can utilize this pointer to send message on the external channel using
// ExternalStateEvent()
var gFSM *FSM

func main() {
	c := make(chan bool)
	go gFSM.RunThread()
	fmt.Println("fsm demo")
	done := <-c
	fmt.Printf("done chan got %v \n", done)
}

/// STATE HANDLERS ///

// StartEntry handles state = start, event= entry
func StartEntry(f *FSM) (FSMStateType, error) {
	fmt.Println("startEntryHandler")
	time.Sleep(10 * time.Second)
	f.InternalStateEvent(FSMInitState, fsmEvtEntry)
	return FSMStartState, nil
}

// StartAbort handles state= start, event=abort
func StartAbort(f *FSM) (FSMStateType, error) {
	fmt.Println("startAbortHandler")
	time.Sleep(10 * time.Second)
	return FSMStartState, nil
}

// InitEntry handles state=init, event=entry
func InitEntry(f *FSM) (FSMStateType, error) {
	fmt.Println("initEntryHandler")
	time.Sleep(10 * time.Second)
	f.InternalStateEvent(FSMStartState, fsmEvtEntry)
	return FSMStartState, nil
}

// InitAbort handles state=init, event=abort
func InitAbort(f *FSM) (FSMStateType, error) {
	fmt.Println("initAbortHandler")
	time.Sleep(10 * time.Second)
	return FSMStartState, nil
}
