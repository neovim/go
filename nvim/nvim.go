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

//go:generate go run api_tool.go -generate api.go

var embedProcAttr *syscall.SysProcAttr

// Nvim represents a remote instance of Nvim. It is safe to call Nvim methods
// concurrently.
type Nvim struct {
	ep *rpc.Endpoint

	// cmd is the child process, if any.
	cmd         *exec.Cmd
	serveCh     chan error
	channelID   int
	channelIDMu sync.Mutex

	// readMu prevents concurrent calls to read on the child process stdout pipe and
	// calls to cmd.Wait().
	readMu sync.Mutex
}

// Serve serves incoming mesages from the peer. Serve blocks until Nvim
// disconnects or there is an error.
//
// By default, the NewChildProcess and Dial functions start a goroutine to run Serve().
// Callers of the low-level New function are responsible for running Serve().
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
	ctx     context.Context
	logf    func(string, ...interface{})
	command string
	dir     string
	args    []string
	env     []string
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
//
// Deprecated: Use ChildProcessOption instead.
type EmbedOptions struct {
	// Logf log function for rpc.WithLogf.
	Logf func(string, ...interface{})

	// Dir specifies the working directory of the command. The working
	// directory in the current process is used if Dir is "".
	Dir string

	// Path is the path of the command to run. If Path = "", then
	// StartEmbeddedNvim searches for "nvim" on $PATH.
	Path string

	// Args specifies the command line arguments. Do not include the program
	// name (the first argument) or the --embed option.
	Args []string

	// Env specifies the environment of the Nvim process. The current process
	// environment is used if Env is nil.
	Env []string
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
		ChannelID int         `msgpack:",array"`
		Info      interface{} `msgpack:"-"`
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
	err     error
	ep      *rpc.Endpoint
	enc     *msgpack.Encoder
	sms     []string
	results []interface{}
	buf     bytes.Buffer
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

// emptyArgs represents a empty interface slice which use to empty args.
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

// batchArg represents a batch call arguments.
type batchArg struct {
	n int
	p []byte
}

// compile time check whether the batchArg implements msgpack.Marshaler interface.
var _ msgpack.Marshaler = (*batchArg)(nil)

// MarshalMsgPack implements msgpack.Marshaler.
func (a *batchArg) MarshalMsgPack(enc *msgpack.Encoder) error {
	enc.PackArrayLen(int64(a.n))
	return enc.PackRaw(a.p)
}

// BatchError represents an error from a API function call in a Batch.
type BatchError struct {
	// Err is the error.
	Err error

	// Index is a zero-based index of the function call which resulted in the
	// error.
	Index int
}

// Error implements the error interface.
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

// Error implements the error interface.
func (el ErrorList) Error() string {
	return el[0].Error()
}

// Request makes a any RPC request.
func (v *Nvim) Request(procedure string, result interface{}, args ...interface{}) error {
	return v.call(procedure, result, args...)
}

// Request makes a any RPC request atomically as a part of batch request.
func (b *Batch) Request(procedure string, result interface{}, args ...interface{}) {
	b.call(procedure, result, args...)
}

// Call calls a VimL function with the given arguments.
//
// Fails with VimL error, does not update "v:errmsg".
//
// fn is Function to call.
//
// args is Function arguments packed in an Array.
//
// result result of the function call.
func (v *Nvim) Call(fname string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_call_function", result, fname, args)
}

// Call calls a VimL function with the given arguments.
//
// Fails with VimL error, does not update "v:errmsg".
//
// fn is Function to call.
//
// args is function arguments packed in an array.
//
// result is result of the function call.
func (b *Batch) Call(fname string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_call_function", result, fname, args)
}

// CallDict calls a VimL dictionary function with the given arguments.
//
// Fails with VimL error, does not update "v:errmsg".
//
// dict is dictionary, or string evaluating to a VimL "self" dict.
//
// fn is name of the function defined on the VimL dict.
//
// args is function arguments packed in an array.
//
// result is result of the function call.
func (v *Nvim) CallDict(dict []interface{}, fname string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_call_dict_function", result, fname, dict, args)
}

// CallDict calls a VimL dictionary function with the given arguments.
//
// Fails with VimL error, does not update "v:errmsg".
//
// dict is dictionary, or string evaluating to a VimL "self" dict.
//
// fn is name of the function defined on the VimL dict.
//
// args is Function arguments packed in an Array.
//
// result is result of the function call.
func (b *Batch) CallDict(dict []interface{}, fname string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_call_dict_function", result, fname, dict, args)
}

// ExecLua execute Lua code.
//
// Parameters are available as `...` inside the chunk. The chunk can return a value.
//
// Only statements are executed. To evaluate an expression, prefix it
// with `return` is  "return my_function(...)".
//
// code is Lua code to execute.
//
// args is arguments to the code.
//
// The returned result value of Lua code if present or nil.
func (v *Nvim) ExecLua(code string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_exec_lua", result, code, args)
}

// ExecLua execute Lua code.
//
// Parameters are available as `...` inside the chunk. The chunk can return a value.
//
// Only statements are executed. To evaluate an expression, prefix it
// with `return` is  "return my_function(...)".
//
// code is Lua code to execute.
//
// args is arguments to the code.
//
// The returned result value of Lua code if present or nil.
func (b *Batch) ExecLua(code string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_exec_lua", result, code, args)
}

// ExecuteLua executes a Lua block.
//
// Deprecated: Use ExecLua() instead.
func (v *Nvim) ExecuteLua(code string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("nvim_execute_lua", result, code, args)
}

// ExecuteLua executes a Lua block.
//
// Deprecated: Use ExecLua() instead.
func (b *Batch) ExecuteLua(code string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	b.call("nvim_execute_lua", result, code, args)
}

// Notify the user with a message.
//
// Relays the call to vim.notify. By default forwards your message in the
// echo area but can be overriden to trigger desktop notifications.
//
// msg is message to display to the user.
//
// logLevel is the LogLevel.
//
// opts is reserved for future use.
func (v *Nvim) Notify(msg string, logLevel LogLevel, opts map[string]interface{}) error {
	if logLevel == LogErrorLevel {
		return v.WritelnErr(msg)
	}

	chunks := []TextChunk{
		{
			Text: msg,
		},
	}
	return v.Echo(chunks, true, opts)
}

// Notify the user with a message.
//
// Relays the call to vim.notify. By default forwards your message in the
// echo area but can be overriden to trigger desktop notifications.
//
// msg is message to display to the user.
//
// logLevel is the LogLevel.
//
// opts is reserved for future use.
func (b *Batch) Notify(msg string, logLevel LogLevel, opts map[string]interface{}) {
	if logLevel == LogErrorLevel {
		b.WritelnErr(msg)
		return
	}

	chunks := []TextChunk{
		{
			Text: msg,
		},
	}
	b.Echo(chunks, true, opts)
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
