// Copyright 2016 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nvim

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/neovim/go-client/msgpack"
	"github.com/neovim/go-client/msgpack/rpc"
)

//go:generate go run apitool.go -generate apiimp.go

var embedProcAttr *syscall.SysProcAttr

// Nvim represents a remote instance of Nvim. It is safe to call Nvim methods
// concurrently.
type Nvim struct {
	ep *rpc.Endpoint

	channelIDMu sync.Mutex
	channelID   int

	// cmd is the child process, if any.
	cmd *exec.Cmd

	serveCh chan error

	// readMu prevents concurrent calls to read on the child process stdout pipe and
	// calls to cmd.Wait().
	readMu sync.Mutex
}

// Serve serves incoming mesages from the peer. Serve blocks until Nvim
// disconnects or there is an error.
//
// By default, the NewChildProcess and Dial functions start a goroutine to run
// Serve(). Callers of the low-level New function are responsible for running
// Serve().
func (v *Nvim) Serve() error {
	v.readMu.Lock()
	defer v.readMu.Unlock()
	return v.ep.Serve()
}

func (v *Nvim) startServe() {
	v.serveCh = make(chan error, 1)
	go func() {
		v.serveCh <- v.Serve()
		close(v.serveCh)
	}()
}

// Close releases the resources used the client.
func (v *Nvim) Close() error {

	if v.cmd != nil && v.cmd.Process != nil {
		// The child process should exit cleanly on call to v.ep.Close(). Kill
		// the process if it does not exit as expected.
		t := time.AfterFunc(10*time.Second, func() { v.cmd.Process.Kill() })
		defer t.Stop()
	}

	err := v.ep.Close()

	if v.cmd != nil {
		v.readMu.Lock()
		defer v.readMu.Unlock()

		errWait := v.cmd.Wait()
		if err == nil {
			err = errWait
		}
	}

	if v.serveCh != nil {
		var errServe error
		select {
		case errServe = <-v.serveCh:
		case <-time.After(10 * time.Second):
			errServe = errors.New("nvim: Serve did not exit")
		}
		if err == nil {
			err = errServe
		}
	}

	return err
}

// New creates an Nvim client. When connecting to Nvim over stdio, use stdin as
// r and stdout as w and c, When connecting to Nvim over a network connection,
// use the connection for r, w and c.
//
// The application must call Serve() to handle RPC requests and responses.
//
// New is a low-level function. Most applications should use NewChildProcess,
// Dial or the ./plugin package.
//
//  :help rpc-connecting
func New(r io.Reader, w io.Writer, c io.Closer, logf func(string, ...interface{})) (*Nvim, error) {
	ep, err := rpc.NewEndpoint(r, w, c, rpc.WithLogf(logf), withExtensions())
	if err != nil {
		return nil, err
	}
	return &Nvim{ep: ep}, nil
}

// ChildProcessOption specifies an option for creating a child process.
type ChildProcessOption struct {
	f func(*childProcessOptions)
}

type childProcessOptions struct {
	args    []string
	command string
	ctx     context.Context
	dir     string
	env     []string
	logf    func(string, ...interface{})
	serve   bool
}

// ChildProcessArgs specifies the command line arguments. The application must
// include the --embed flag or other flags that cause Nvim to use stdin/stdout
// as a MsgPack RPC channel.
func ChildProcessArgs(args ...string) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.args = args
	}}
}

// ChildProcessCommand specifies the command to run. NewChildProcess runs
// "nvim" by default.
func ChildProcessCommand(command string) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.command = command
	}}
}

// ChildProcessContext specifies the context to use when starting the command.
// The background context is used by defaullt.
func ChildProcessContext(ctx context.Context) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.ctx = ctx
	}}
}

// ChildProcessDir specifies the working directory for the process. The current
// working directory is used by default.
func ChildProcessDir(dir string) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.dir = dir
	}}
}

// ChildProcessEnv specifies the environment for the child process. The current
// process environment is used by default.
func ChildProcessEnv(env []string) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.env = env
	}}
}

// ChildProcessServe specifies whether Server should be run in a goroutine.
// The default is to run Serve().
func ChildProcessServe(serve bool) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.serve = serve
	}}
}

// ChildProcessLogf specifies function for logging output. The log.Printf
// function is used by default.
func ChildProcessLogf(logf func(string, ...interface{})) ChildProcessOption {
	return ChildProcessOption{func(cpos *childProcessOptions) {
		cpos.logf = logf
	}}
}

// NewChildProcess returns a client connected to stdin and stdout of a new
// child process.
func NewChildProcess(options ...ChildProcessOption) (*Nvim, error) {

	cpos := &childProcessOptions{
		serve:   true,
		logf:    log.Printf,
		command: "nvim",
		ctx:     context.Background(),
	}
	for _, cpo := range options {
		cpo.f(cpos)
	}

	cmd := exec.CommandContext(cpos.ctx, cpos.command, cpos.args...)
	cmd.Env = cpos.env
	cmd.Dir = cpos.dir
	cmd.SysProcAttr = embedProcAttr

	inw, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	outr, err := cmd.StdoutPipe()
	if err != nil {
		inw.Close()
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	v, _ := New(outr, inw, inw, cpos.logf)
	v.cmd = cmd

	if cpos.serve {
		v.startServe()
	}

	return v, nil
}

// EmbedOptions specifies options for starting an embedded instance of Nvim.
type EmbedOptions struct {
	// Args specifies the command line arguments. Do not include the program
	// name (the first argument) or the --embed option.
	Args []string

	// Dir specifies the working directory of the command. The working
	// directory in the current process is used if Dir is "".
	Dir string

	// Env specifies the environment of the Nvim process. The current process
	// environment is used if Env is nil.
	Env []string

	// Path is the path of the command to run. If Path = "", then
	// StartEmbeddedNvim searches for "nvim" on $PATH.
	Path string

	Logf func(string, ...interface{})
}

// NewEmbedded starts an embedded instance of Nvim using the specified options.
//
// The application must call Serve() to handle RPC requests and responses.
//
// Deprecated: Use NewChildProcess instead.
func NewEmbedded(options *EmbedOptions) (*Nvim, error) {
	if options == nil {
		options = &EmbedOptions{}
	}
	path := options.Path
	if path == "" {
		path = "nvim"
	}

	return NewChildProcess(
		ChildProcessArgs(append([]string{"--embed"}, options.Args...)...),
		ChildProcessCommand(path),
		ChildProcessEnv(options.Env),
		ChildProcessDir(options.Dir),
		ChildProcessServe(false))
}

// DialOption specifies an option for dialing to Nvim.
type DialOption struct {
	f func(*dialOptions)
}

type dialOptions struct {
	ctx     context.Context
	logf    func(string, ...interface{})
	netDial func(ctx context.Context, network, address string) (net.Conn, error)
	serve   bool
}

// DialContext specifies the context to use when starting the command.
// The background context is used by default.
func DialContext(ctx context.Context) DialOption {
	return DialOption{func(dos *dialOptions) {
		dos.ctx = ctx
	}}
}

// DialNetDial specifies a function used to dial a network connection. A
// default net.Dialer DialContext method is used by default.
func DialNetDial(f func(ctx context.Context, network, address string) (net.Conn, error)) DialOption {
	return DialOption{func(dos *dialOptions) {
		dos.netDial = f
	}}
}

// DialServe specifies whether Server should be run in a goroutine.
// The default is to run Serve().
func DialServe(serve bool) DialOption {
	return DialOption{func(dos *dialOptions) {
		dos.serve = serve
	}}
}

// DialLogf specifies function for logging output. The log.Printf function is used by default.
func DialLogf(logf func(string, ...interface{})) DialOption {
	return DialOption{func(dos *dialOptions) {
		dos.logf = logf
	}}
}

// Dial dials an Nvim instance given an address in the format used by
// $NVIM_LISTEN_ADDRESS.
//
//  :help rpc-connecting
//  :help $NVIM_LISTEN_ADDRESS
func Dial(address string, options ...DialOption) (*Nvim, error) {
	var d net.Dialer
	dos := &dialOptions{
		ctx:     context.Background(),
		logf:    log.Printf,
		netDial: d.DialContext,
		serve:   true,
	}

	for _, do := range options {
		do.f(dos)
	}

	network := "unix"
	if strings.Contains(address, ":") {
		network = "tcp"
	}

	c, err := dos.netDial(dos.ctx, network, address)
	if err != nil {
		return nil, err
	}

	v, err := New(c, c, c, dos.logf)
	if err != nil {
		c.Close()
		return nil, err
	}

	if dos.serve {
		v.startServe()
	}
	return v, err
}

// RegisterHandler registers fn as a MessagePack RPC handler for the named
// method. The function signature for fn is one of
//
//  func([v *nvim.Nvim,] {args}) ({resultType}, error)
//  func([v *nvim.Nvim,] {args}) error
//  func([v *nvim.Nvim,] {args})
//
// where {args} is zero or more arguments and {resultType} is the type of a
// return value. Call the handler from Nvim using the rpcnotify and rpcrequest
// functions:
//
//  :help rpcrequest()
//  :help rpcnotify()
//
// Plugin applications should use the Handler* methods in the ./plugin package
// to register handlers instead of this method.
func (v *Nvim) RegisterHandler(method string, fn interface{}) error {
	var args []interface{}
	t := reflect.TypeOf(fn)
	if t.Kind() == reflect.Func && t.NumIn() > 0 && t.In(0) == reflect.TypeOf(v) {
		args = append(args, v)
	}
	return v.ep.Register(method, fn, args...)
}

// ChannelID returns Nvim's channel id for this client.
func (v *Nvim) ChannelID() int {
	v.channelIDMu.Lock()
	defer v.channelIDMu.Unlock()
	if v.channelID != 0 {
		return v.channelID
	}
	var info struct {
		ChannelID int `msgpack:",array"`
		Info      interface{}
	}
	if err := v.ep.Call("nvim_get_api_info", &info); err != nil {
		// TODO: log error and exit process?
	}
	v.channelID = info.ChannelID
	return v.channelID
}

func (v *Nvim) call(sm string, result interface{}, args ...interface{}) error {
	return fixError(sm, v.ep.Call(sm, result, args...))
}

// NewBatch creates a new batch.
func (v *Nvim) NewBatch() *Batch {
	b := &Batch{ep: v.ep}
	b.enc = msgpack.NewEncoder(&b.buf)
	return b
}

// Batch collects API function calls and executes them atomically.
//
// The function calls in the batch are executed without processing requests
// from other clients, redrawing or allowing user interaction in between.
// Functions that could fire autocommands or do event processing still might do
// so. For instance invoking the :sleep command might call timer callbacks.
//
// Call the Execute() method to execute the commands in the batch. Result
// parameters in the API function calls are set in the call to Execute.  If an
// API function call fails, all results proceeding the call are set and a
// *BatchError is returned.
//
// A Batch does not support concurrent calls by the application.
type Batch struct {
	ep      *rpc.Endpoint
	buf     bytes.Buffer
	enc     *msgpack.Encoder
	sms     []string
	results []interface{}
	err     error
}

// Execute executes the API function calls in the batch.
func (b *Batch) Execute() error {
	defer func() {
		b.buf.Reset()
		b.sms = b.sms[:0]
		b.results = b.results[:0]
		b.err = nil
	}()

	if b.err != nil {
		return b.err
	}

	result := struct {
		Results []interface{} `msgpack:",array"`
		Error   *struct {
			Index   int `msgpack:",array"`
			Type    int
			Message string
		}
	}{
		b.results,
		nil,
	}

	err := b.ep.Call("nvim_call_atomic", &result, &batchArg{n: len(b.sms), p: b.buf.Bytes()})
	if err != nil {
		return err
	}

	e := result.Error
	if e == nil {
		return nil
	}

	if e.Index < 0 || e.Index >= len(b.sms) ||
		(e.Type != exceptionError && e.Type != validationError) {
		return fmt.Errorf("nvim:nvim_call_atomic %d %d %s", e.Index, e.Type, e.Message)
	}
	errorType := "exception"
	if e.Type == validationError {
		errorType = "validation"
	}
	return &BatchError{
		Index: e.Index,
		Err:   fmt.Errorf("nvim:%s %s: %s", b.sms[e.Index], errorType, e.Message),
	}
}

var emptyArgs = []interface{}{}

func (b *Batch) call(sm string, result interface{}, args ...interface{}) {
	if b.err != nil {
		return
	}
	if args == nil {
		args = emptyArgs
	}
	b.sms = append(b.sms, sm)
	b.results = append(b.results, result)
	b.enc.PackArrayLen(2)
	b.enc.PackString(sm)
	b.err = b.enc.Encode(args)
}

type batchArg struct {
	n int
	p []byte
}

func (a *batchArg) MarshalMsgPack(enc *msgpack.Encoder) error {
	enc.PackArrayLen(int64(a.n))
	return enc.PackRaw(a.p)
}

// BatchError represents an error from a API function call in a Batch.
type BatchError struct {
	// Index is a zero-based index of the function call which resulted in the
	// error.
	Index int

	// Err is the error.
	Err error
}

func (e *BatchError) Error() string {
	return e.Err.Error()
}

func fixError(sm string, err error) error {
	if e, ok := err.(rpc.Error); ok {
		if a, ok := e.Value.([]interface{}); ok && len(a) == 2 {
			switch a[0] {
			case int64(exceptionError), uint64(exceptionError):
				return fmt.Errorf("nvim:%s exception: %v", sm, a[1])
			case int64(validationError), uint64(validationError):
				return fmt.Errorf("nvim:%s validation: %v", sm, a[1])
			}
		}
	}
	return err
}

// ErrorList is a list of errors.
type ErrorList []error

func (el ErrorList) Error() string {
	return el[0].Error()
}

// Call calls a vimscript function.
func (v *Nvim) Call(fname string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_call_function", result, fname, args)
}

// Call calls a vimscript function.
func (b *Batch) Call(fname string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_call_function", result, fname, args)
}

// CallDict calls a vimscript Dictionary function.
func (v *Nvim) CallDict(dict []interface{}, fname string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_call_dict_function", result, fname, dict, args)
}

// CallDict calls a vimscript Dictionary function.
func (b *Batch) CallDict(dict []interface{}, fname string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_call_dict_function", result, fname, dict, args)
}

// ExecuteLua executes a Lua block.
func (v *Nvim) ExecuteLua(code string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_execute_lua", result, code, args)
}

// ExecuteLua executes a Lua block.
func (b *Batch) ExecuteLua(code string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_execute_lua", result, code, args)
}

// decodeExt decodes a MsgPack encoded number to go int value.
func decodeExt(p []byte) (int, error) {
	switch {
	case len(p) == 1 && p[0] <= 0x7f:
		return int(p[0]), nil
	case len(p) == 2 && p[0] == 0xcc:
		return int(p[1]), nil
	case len(p) == 3 && p[0] == 0xcd:
		return int(uint16(p[2]) | uint16(p[1])<<8), nil
	case len(p) == 5 && p[0] == 0xce:
		return int(uint32(p[4]) | uint32(p[3])<<8 | uint32(p[2])<<16 | uint32(p[1])<<24), nil
	case len(p) == 2 && p[0] == 0xd0:
		return int(int8(p[1])), nil
	case len(p) == 3 && p[0] == 0xd1:
		return int(int16(uint16(p[2]) | uint16(p[1])<<8)), nil
	case len(p) == 5 && p[0] == 0xd2:
		return int(int32(uint32(p[4]) | uint32(p[3])<<8 | uint32(p[2])<<16 | uint32(p[1])<<24)), nil
	case len(p) == 1 && p[0] >= 0xe0:
		return int(int8(p[0])), nil
	default:
		return 0, fmt.Errorf("go-client/nvim: error decoding extension bytes %x", p)
	}
}

// encodeExt encodes n to MsgPack format.
func encodeExt(n int) []byte {
	return []byte{0xd2, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

func unmarshalExt(dec *msgpack.Decoder, id int, v interface{}) (int, error) {
	if dec.Type() != msgpack.Extension || dec.Extension() != id {
		err := &msgpack.DecodeConvertError{
			SrcType:  dec.Type(),
			DestType: reflect.TypeOf(v).Elem(),
		}
		dec.Skip()
		return 0, err
	}
	return decodeExt(dec.BytesNoCopy())
}

type Mode struct {
	// Mode is the current mode.
	Mode string `msgpack:"mode"`

	// Blocking is true if Nvim is waiting for input.
	Blocking bool `msgpack:"blocking"`
}

type HLAttrs struct {
	Bold       bool `msgpack:"bold,omitempty"`
	Underline  bool `msgpack:"underline,omitempty"`
	Undercurl  bool `msgpack:"undercurl,omitempty"`
	Italic     bool `msgpack:"italic,omitempty"`
	Reverse    bool `msgpack:"reverse,omitempty"`
	Foreground int  `msgpack:"foreground,omitempty" empty:"-1"`
	Background int  `msgpack:"background,omitempty" empty:"-1"`
	Special    int  `msgpack:"special,omitempty" empty:"-1"`
}

type Mapping struct {
	// LHS is the {lhs} of the mapping.
	LHS string `msgpack:"lhs"`

	// RHS is the {hrs} of the mapping as typed.
	RHS string `msgpack:"rhs"`

	// Silent is 1 for a |:map-silent| mapping, else 0.
	Silent int `msgpack:"silent"`

	// Noremap is 1 if the {rhs} of the mapping is not remappable.
	NoRemap int `msgpack:"noremap"`

	// Expr is  1 for an expression mapping.
	Expr int `msgpack:"expr"`

	// Buffer for a local mapping.
	Buffer int `msgpack:"buffer"`

	// SID is the script local ID, used for <sid> mappings.
	SID int `msgpack:"sid"`

	// Nowait is 1 if map does not wait for other, longer mappings.
	NoWait int `msgpack:"nowait"`

	// Mode specifies modes for which the mapping is defined.
	Mode string `msgpack:"string"`
}

type Channel struct {
	// ID is channel id.
	ID int `msgpack:"id,omitempty"`
	// Stream is the stream underlying the channel.
	Stream string `msgpack:"stream,omitempty"`
	// Mode is the how data received on the channel is interpreted.
	Mode string `msgpack:"mode,omitempty"`
	// Pty is the name of pseudoterminal, if one is used (optional).
	Pty string `msgpack:"pty,omitempty"`
	// Buffer is the buffer with connected terminal instance (optional).
	Buffer string `msgpack:"buffer,omitempty"`
	// Client is the information about the client on the other end of the RPC channel, if it has added it using nvim_set_client_info (optional).
	Client *Client `msgpack:"client,omitempty"`
}

type Client struct {
	// Name is short name for the connected client.
	Name string `msgpack:"name,omitempty"`
	// Version describes the version, with the following possible keys (all optional).
	Version map[string]interface{} `msgpack:"version,omitempty"`
	// Type is client type. A client library should use "remote" if the library user hasn't specified other value.
	Type string `msgpack:"type,omitempty"`
	// Methods builtin methods in the client.
	Methods map[string]interface{} `msgpack:"methods,omitempty"`
	// Attributes is informal attributes describing the client.
	Attributes map[string]interface{} `msgpack:"attributes,omitempty"`
}

type Process struct {
	// Name is the name of process command.
	Name string `msgpack:"name,omitempty"`
	// PID is the process ID.
	PID int `msgpack:"pid,omitempty"`
	// PPID is the parent process ID.
	PPID int `msgpack:"ppid,omitempty"`
}

type UI struct {
	// Height requested height of the UI
	Height int `msgpack:"height,omitempty"`
	// Width requested width of the UI
	Width int `msgpack:"width,omitempty"`
	// RGB whether the UI uses rgb colors (false implies cterm colors)
	RGB bool `msgpack:"rgb,omitempty"`
	// ExtPopupmenu externalize the popupmenu.
	ExtPopupmenu bool `msgpack:"ext_popupmenu,omitempty"`
	// ExtTabline externalize the tabline.
	ExtTabline bool `msgpack:"ext_tabline,omitempty"`
	// ExtCmdline externalize the cmdline.
	ExtCmdline bool `msgpack:"ext_cmdline,omitempty"`
	// ExtWildmenu externalize the wildmenu.
	ExtWildmenu bool `msgpack:"ext_wildmenu,omitempty"`
	// ExtNewgrid use new revision of the grid events.
	ExtNewgrid bool `msgpack:"ext_newgrid,omitempty"`
	// ExtHlstate use detailed highlight state.
	ExtHlstate bool `msgpack:"ext_hlstate,omitempty"`
	// ChannelID channel id of remote UI (not present for TUI)
	ChannelID int `msgpack:"chan,omitempty"`
}
