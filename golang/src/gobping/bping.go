package main

// #cgo LDFLAGS: -lbp -lici
// #include <bp.h>
// #include <platform.h>
// #include <zco.h>
// #include <sdrxn.h>
// #include <sdrmgt.h>
// #include <sdrstring.h>
// #include <sdrtable.h>
//
// vast min(vast x, vast y) {
//	(x < y) ? x : y;
// }
import "C"

import (
	"errors"
	"flag"
	"fmt"
	"sync"
	"unsafe"
)

var (
	sEid      = flag.String("seid", "", "source eid")
	dEid      = flag.String("deid", "", "destination EID")
	ttl       = flag.Int("ttl", 3600, "TTL")
	priority  = flag.String("priority", "0", "priority")
	verbosity bool
)

type SafeSdr struct {
	mu  sync.Mutex
	sdr C.Sdr
}

type Watch struct {
	wg      sync.WaitGroup
	toClose []chan bool
	close   chan bool
}

func NewWatch() *Watch {
	return &Watch{}
}

func (w *Watch) Add() chan bool {
	w.wg.Add(1)
	c := make(chan bool)
	w.toClose = append(w.toClose, c)
	return c
}

func (w *Watch) Start() {
	for {
		select {
		case <-w.close:
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

func BpReceiveResponse(recvsap C.BpSAP, safeSdr *SafeSdr, w *Watch) {

	close := w.Add()
	received := make(chan bool)

	var dlv C.BpDelivery
	var reader C.ZcoReader
	var buffer = make([]byte, BPING_PAYLOAD_MAX_LEN)

	var cBuffer *C.char

OuterLoop:
	for {
		if int(C.bp_receive(recvsap, &dlv, C.BP_BLOCKING)) >= 0 {
			received <- true
		}

		select {
		case <-close:
			break OuterLoop

		case <-received:
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
			bytesToRead := C.min(contentLength, C.longlong(len(buffer)-1))
			C.zco_start_receiving(dlv.adu, &reader)
			_ = int(C.sdr_begin_xn(safeSdr.sdr))
			cBuffer = C.CString(string(buffer))
			_ = C.zco_receive_source(safeSdr.sdr, &reader, bytesToRead, cBuffer)
			C.bp_release_delivery(&dlv, 1)

			safeSdr.mu.Unlock()

			fmt.Println(C.GoString(cBuffer))
		}
	}
	C.free(unsafe.Pointer(cBuffer))
}

func BpSendRequest(payload string, xmitsap C.BpSAP, recvsap C.BpSAP, safeSdr *SafeSdr, w *Watch) {
	close := w.Add()

OuterLoop:
	for {
		select {
			case <-close:
				break OuterLoop

			default:
				safeSdr.mu.Lock()

				//C.bp_interrupt(recvsap)

				pchar := (*C.uchar)(unsafe.Pointer(priority))
				pint := (*C.int)(unsafe.Pointer(priority))
				cBuffer := C.CString(string(payload))
				bundleMessage := C.Sdr_malloc(cBuffer, 0, safeSdr.sdr, C.ulong(len(payload)))
				bundleZco := C.ionCreateZco(C.ZcoSdrSource, bundleMessage, 0, C.longlong(len(payload)), *pchar, 0, C.ZcoOutbound, nil)


				if(bundleZco == 0) {
					//C.bp_interrupt(recvsap)
					w.close <- true
				}

				if(C.bp_send(xmitsap, C.CString(*dEid), nil, C.int(*ttl), *pint, 0, 0, 0, nil, bundleZco, nil) <= 0) {
					//C.bp_interrupt(recvsap)
					w.close <- true
				}

				safeSdr.mu.Unlock()
				w.close <- true
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

	safeSdr.sdr = C.bp_get_sdr()

	go BpReceiveResponse(recvsap, &safeSdr, w)
	go BpSendRequest("blah", xmitsap, recvsap, &safeSdr, w)

	w.Start()

	C.bp_close(xmitsap)
	C.bp_close(recvsap)
	BpDetach()
	C.free(unsafe.Pointer(cs))
}
