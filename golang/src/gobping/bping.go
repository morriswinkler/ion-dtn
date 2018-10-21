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
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	sEid      = flag.String("seid", "ipn:1.1", "source eid")
	dEid      = flag.String("deid", "ipn:2.1", "destination EID")
	ttl       = flag.Int("ttl", 3600, "TTL")
	priority  = flag.String("priority", "3", "priority")
	verbosity bool

	watcher = NewWatch()
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
	return &Watch{close: make(chan bool)}
}

func (w *Watch) Add() chan bool {
	w.wg.Add(1)
	c := make(chan bool)
	w.toClose = append(w.toClose, c)
	return c
}

func (w *Watch) Start() {
	<-w.close
	for i := range w.toClose {
		fmt.Printf("close gopher %d \n", i)
		w.toClose[i] <- true
	}
	fmt.Println("close gopher wait")
	w.wg.Wait()
	fmt.Println("close gopher wait done")
}

func init() {
	flag.BoolVar(&verbosity, "v", false, "enable verbosity")

	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		sig := <-c
		fmt.Println()
		fmt.Println(sig)
		watcher.close <- true
	}()
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

	defer w.wg.Done()
	watchChan := w.Add()

	received := make(chan bool)

	var dlv C.BpDelivery
	var reader C.ZcoReader
	var buffer = make([]byte, BPING_PAYLOAD_MAX_LEN)

	var cBuffer *C.char
	defer func() {
		C.free(unsafe.Pointer(cBuffer))
	}()

	go func() {
		for {
			if int(C.bp_receive(recvsap, &dlv, C.BP_BLOCKING)) >= 0 {
				received <- true
			}
		}
	}()

	for {
		fmt.Println("BpReceiveResponse")
		select {
		case <-watchChan:
			fmt.Println("close BpReceiveResponse")
			C.bp_release_delivery(&dlv, 1)
			return

		case <-received:
			fmt.Println("Received ....")

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
}

func BpSendRequest(payload string, xmitsap C.BpSAP, recvsap C.BpSAP, safeSdr *SafeSdr, w *Watch) {
	defer w.wg.Done()
	watchChan := w.Add()

	defer C.bp_interrupt(recvsap)

	ticker := time.NewTicker(time.Second)

	for {
		fmt.Println("BpSendRequest")
		select {
		case <-watchChan:
			fmt.Println("close BpSendRequest")
			return
		case <-ticker.C:
			fmt.Println("BpSendRequest default")

			safeSdr.mu.Lock()

			fmt.Println("BpSendRequest lock")

			var pchar2 C.uchar = 0
			var pint2 C.int = 0
			cBuffer := C.CString(string(payload))

			fmt.Println("BpSendRequest sdr_begin_xn before")

			r := int(C.sdr_begin_xn(safeSdr.sdr))
			if r <= 0 {
				fmt.Printf("sdr_begin_xn returned: %d\n", r)
				w.close <- true
			}

			fmt.Println("BpSendRequest sdr_begin_xn after ")

			bundleMessage := C.Sdr_malloc(cBuffer, 0, safeSdr.sdr, C.ulong(len(payload)))

			if bundleMessage != 0 {
				C.Sdr_write(cBuffer, 0, safeSdr.sdr, bundleMessage, cBuffer, C.ulong(len(payload)))
			}

			r = int(C.sdr_end_xn(safeSdr.sdr))
			if r > 1 {
				fmt.Printf("sdr_end_xn returned: %d\n", r)
				w.close <- true
			}

			fmt.Println("BpSendRequest center")

			bundleZco := C.ionCreateZco(C.ZcoSdrSource, bundleMessage, 0, C.longlong(len(payload)), pchar2, 0, C.ZcoOutbound, nil)

			if bundleZco == 0 {
				w.close <- true
			}

			if C.bp_send(xmitsap, C.CString(*dEid), nil, C.int(*ttl), pint2, 0, 0, 0, nil, bundleZco, nil) <= 0 {
				w.close <- true
			}

			fmt.Println("BpSendRequest bottom")

			safeSdr.mu.Unlock()
		}
	}
}

func main() {

	flag.Parse()
	cs := C.CString(*sEid)

	var xmitsap, recvsap C.BpSAP
	var safeSdr SafeSdr

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

	go BpReceiveResponse(recvsap, &safeSdr, watcher)
	go BpSendRequest("blah", xmitsap, recvsap, &safeSdr, watcher)

	fmt.Println("1")

	watcher.Start()

	fmt.Println("1")

	C.bp_close(xmitsap)
	C.bp_close(recvsap)
	BpDetach()
	fmt.Println("BpDetach called")
	C.free(unsafe.Pointer(cs))
}
