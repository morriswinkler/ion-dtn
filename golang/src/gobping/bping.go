package main

// #cgo LDFLAGS: -lbp
// #include <bp.h>
import "C"

import (
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var (
	sEid      = flag.String("seid", "", "source eid")
	verbosity bool
)

type SafeSdr struct {
	mu  sync.Mutex
	sdr C.Sdr
}

type Watch struct {
	wg sync.WaitGroup
	toClose []chan bool
	close chan bool
}

func NewWatch() *Watch {
	return &Watch{}
}

func (w *Watch) Add() chan bool{
	w.wg.Add(1)
	c := make(chan bool)
	w.toClose = append(w.toClose, c)
	return c
}

func (w *Watch) Start() {
	for {
		select {
			case <- w.close;
			for i := range w.toClose {
				w.toClose[i] <- true
			}
			w.wg.Wait()
		}
	}
}

func init() {
	flag.BoolVar(&verbosity, "v", false, "enable verbosity")
}

func BpAttach() error {
	if int(C.bp_attach()) < 0 {
		return errors.New("bp_attach failed")
	}
	return nil
}

func BpOpenSource(cs *C.char, sap *C.BpSAP) error {
	if int(C.bp_open_source(cs, sap, 0)) < 0 {
		return errors.New("bp_open_source failed")
	}
	return nil
}

func BpOpen(cs *C.char, sap *C.BpSAP) error {
	if int(C.bp_open(cs, sap)) < 0 {
		return errors.New("bp_open_source failed")
	}
	return nil
}

func BpDetach() {
	C.bp_detach()
}

const BPING_PAYLOAD_MAX_LEN = 65537

func BpReceiveResponse(recvsap C.BpSAP, safeSdr *SafeSdr, w *Watch ) {

	close := w.Add()

	var dlv C.BpDelivery
	var reader C.ZcoReader
	var buffer [BPING_PAYLOAD_MAX_LEN]C.Char
	for {
		select {
		case <-close:
			close <- true
			continue

		case C.bp_receive(recvsap, &dlv, C.BP_BLOCKING) >= 0:
			//now := time.Now()

			if dlv.result == C.BpReceptionInterrupted || dlv.adu == 0 {
				if verbosity {
					fmt.Println("Reception interrupted.\n")
				}
				C.bp_release_delivery(&dlv, 1)
				w.close <- true
			}

			if dlv.result == C.BpEndpointStopped {
				if verbosity {
					fmt.Println("Endpoint stopped.\n")
				}
				C.bp_release_delivery(&dlv, 1)
				w.close <- true
			}

			safeSdr.mu.Lock()

			contentLength := C.zco_source_data_length(safeSdr.sdr, dlv.adu)
			bytesToRead := C.MIN(contentLength, C.sizeof(buffer)-1)
			C.zco_start_receiving(dlv.adu, &reader)
			C.oK(C.sdr_begin_xn(safeSdr.sdr))
			_ = C.zco_receive_source(safeSdr.sdr, &reader, bytesToRead, buffer)
			C.bp_release_delivery(&dlv, 1)

			safeSdr.mu.Unlock()

			fmt.Println(buffer)
		}
	}
}

func main() {

	flag.Parse()
	cs := C.CString(*sEid)

	var xmitsap, recvsap C.BpSAP
	var safeSdr SafeSdr

	w := NewWatch()

	err := BpAttach()
	if err != nil {
		panic(err)
	}

	err = BpOpenSource(cs, &xmitsap)
	if err != nil {
		BpDetach()
		panic(err)
	}

	err = BpOpen(cs, &recvsap)
	if err != nil {
		BpDetach()
		panic(err)
	}

	safeSdr = C.bp_get_sdr()

	go BpReceiveResponse(recvsap, &safeSdr, w)

	go w.Start()

	C.bp_close(xmitsap)
	C.bp_close(recvsap)
	BpDetach()
	C.free(unsafe.Pointer(cs))
}
