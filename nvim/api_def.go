//go:build ignore
// +build ignore

// This file defines the Nvim remote API using Go syntax. Run the 'go generate'
// command to convert this file to the API implementation in api.go.
//
// The go generate command runs the apitool program. Run
//
//  go run api_tool.go --help
//
// to learn more about the apitool program.

package main

// vim.c

// Exec executes Vimscript (multiline block of Ex-commands), like anonymous source.
//
// Unlike Command, this function supports heredocs, script-scope (s:), etc.
//
// When fails with VimL error, does not update "v:errmsg".
func Exec(src string, output bool) (out string) {
	name(nvim_exec)
}

// Command executes an ex-command.
//
// When fails with VimL error, does not update "v:errmsg".
func Command(cmd string) {
	name(nvim_command)
}

// HLByName gets a highlight definition by id.
//
// name is Highlight group name.
//
// rgb is whether the export RGB colors.
//
// The returned highlight is the highlight definition.
func HLByName(name string, rgb bool) (highlight HLAttrs) {
	name(nvim_get_hl_by_name)
	returnPtr()
}

// HLByID gets a highlight definition by name.
//
// hlID is the highlight id as returned by HLIDByName.
//
// rgb is the whether the export RGB colors.
//
// The returned highlight is the highlight definition.
func HLByID(hlID int, rgb bool) (highlight HLAttrs) {
	name(nvim_get_hl_by_id)
	returnPtr()
}

// HLIDByName gets a highlight group by name.
//
// name is the Highlight group name.
//
// The returns hlID is the highlight id.
//
// This function similar to HLByID, but allocates a new ID if not present.
func HLIDByName(name string) (hlID int) {
	name(nvim_get_hl_id_by_name)
}

// SetHighlight sets a highlight group.
//
// nsID is number of namespace for this highlight.
//
// name is highlight group name, like "ErrorMsg".
//
// val is highlight definiton map, like HLByName.
//
// in addition the following keys are also recognized:
//
//  default
// don't override existing definition, like "hi default".
func SetHighlight(nsID int, name string, val *HLAttrs) {
	name(nvim_set_hl)
}

// SetHighlightNameSpace set active namespace for highlights.
//
// nsID is the namespace to activate.
func SetHighlightNameSpace(nsID int) {
	name(nvim__set_hl_ns)
}

// FeedKeys input-keys to Nvim, subject to various quirks controlled by "mode"
// flags. Unlike Input, this is a blocking call.
//
// This function does not fail, but updates "v:errmsg".
//
// If need to input sequences like <C-o> use ReplaceTermcodes to
// replace the termcodes and then pass the resulting string to nvim_feedkeys.
// You'll also want to enable escape_csi.
//
// mode is following character flags:
//
//  m
// Remap keys. This is default.
//
//  n
// Do not remap keys.
//
//  t
// Handle keys as if typed; otherwise they are handled as if coming from a mapping.
// This matters for undo, opening folds, etc.
//
// escapeCSI is whether the escape K_SPECIAL/CSI bytes in keys.
func FeedKeys(keys, mode string, escapeCSI bool) {
	name(nvim_feedkeys)
}

// Input queues raw user-input.
//
// Unlike FeedKeys, this uses a low-level input buffer and the call
// is non-blocking (input is processed asynchronously by the eventloop).
//
// This function does not fail but updates "v:errmsg".
//
// keys is to be typed.
//
// Note: "keycodes" like "<CR>" are translated, so "<" is special. To input a literal "<", send "<LT>".
//
// Note: For mouse events use InputMouse. The pseudokey form "<LeftMouse><col,row>" is deprecated.
//
// The returned written is number of bytes actually written (can be fewer than
// requested if the buffer becomes full).
func Input(keys string) (written int) {
	name(nvim_input)
}

// InputMouse Send mouse event from GUI.
//
// This API is non-blocking. It does not wait on any result, but queues the event to be
// processed soon by the event loop.
//
// The button arg is mouse button. One of
//  "left"
//  "right"
//  "middle"
//  "wheel"
//
// The action is for ordinary buttons. One of
//  "press"
//  "drag"
//  "release"
// For the wheel, One of
//  "up"
//  "down
//  "left"
//  "right"
//
// modifier is string of modifiers each represented by a single char.
// The same specifiers are used as for a key press, except
// that the "-" separator is optional, so "C-A-", "c-a"
// and "CA" can all be used to specify Ctrl+Alt+click.
//
// grid is grid number if the client uses "ui-multigrid", else 0.
//
// row is mouse row-position (zero-based, like redraw events).
//
// col is mouse column-position (zero-based, like redraw events).
func InputMouse(button, action, modifier string, grid, row, col int) {
	name(nvim_input_mouse)
}

// ReplaceTermcodes replaces terminal codes and "keycodes" (<CR>, <Esc>, ...) in a string with
// the internal representation.
//
// str is string to be converted.
//
// fromPart is legacy Vim parameter. Usually true.
//
// doLT is also translate <lt>. Ignored if "special" is false.
//
// special is replace "keycodes", e.g. "<CR>" becomes a "\n" char.
//
// The returned sequences are Nvim's internal representation of keys, for example:
//
//  <esc> -> '\x1b'
//  <cr>  -> '\r'
//  <c-l> -> '\x0c'
//  <up>  -> '\x80ku'
//
// The returned sequences can be used as input to feedkeys.
func ReplaceTermcodes(str string, fromPart, doLT, special bool) (input string) {
	name(nvim_replace_termcodes)
}

// CommandOutput executes a single ex command and returns the output.
//
// Deprecated: Use Exec() instead.
func CommandOutput(cmd string) (out string) {
	name(nvim_command_output)
	deprecatedSince(7)
}

// Eval evaluates a VimL expression.
//
// Dictionaries and Lists are recursively expanded.
//
// Fails with VimL error, does not update "v:errmsg".
//
// expr is VimL expression string.
//
//  :help expression
func Eval(expr string) (result interface{}) {
	name(nvim_eval)
}

// StringWidth calculates the number of display cells occupied by "text".
//
// "<Tab>" counts as one cell.
func StringWidth(s string) (width int) {
	name(nvim_strwidth)
}

// RuntimePaths gets the paths contained in "runtimepath".
func RuntimePaths() (paths []string) {
	name(nvim_list_runtime_paths)
}

// RuntimeFiles find files in runtime directories.
//
// name is can contain wildcards.
//
// For example,
//
//  RuntimeFiles("colors/*.vim", true)
//
// will return all color scheme files.
//
// Always use forward slashes (/) in the search pattern for subdirectories regardless of platform.
//
// It is not an error to not find any files, returned an empty array.
//
// To find a directory, name must end with a forward slash, like
// "rplugin/python/".
// Without the slash it would instead look for an ordinary file called "rplugin/python".
//
// all is whether to return all matches or only the first.
func RuntimeFiles(name string, all bool) (files []string) {
	name(nvim_get_runtime_file)
}

// SetCurrentDirectory changes the global working directory.
func SetCurrentDirectory(dir string) {
	name(nvim_set_current_dir)
}

// CurrentLine gets the current line.
func CurrentLine() (line []byte) {
	name(nvim_get_current_line)
}

// SetCurrentLine sets the current line.
func SetCurrentLine(line []byte) {
	name(nvim_set_current_line)
}

// DeleteCurrentLine deletes the current line.
func DeleteCurrentLine() {
	name(nvim_del_current_line)
}

// Var gets a global (g:) variable.
func Var(name string) (value interface{}) {
	name(nvim_get_var)
}

// SetVar sets a global (g:) variable.
func SetVar(name string, value interface{}) {
	name(nvim_set_var)
}

// DeleteVar removes a global (g:) variable.
func DeleteVar(name string) {
	name(nvim_del_var)
}

// VVar gets a v: variable.
func VVar(name string) (value interface{}) {
	name(nvim_get_vvar)
}

// SetVVar sets a v: variable, if it is not readonly.
func SetVVar(name string, value interface{}) {
	name(nvim_set_vvar)
}

// Option gets an option value string.
func Option(name string) (option interface{}) {
	name(nvim_get_option)
}

// AllOptionsInfo gets the option information for all options.
//
// The dictionary has the full option names as keys and option metadata
// dictionaries as detailed at OptionInfo.
//
// Resulting map has keys:
//
//  name
// Name of the option (like "filetype").
//
//  shortname
// Shortened name of the option (like "ft").
//
//  type
// type of option ("string", "number" or "boolean").
//
//  default
// The default value for the option.
//
//  was_set
// Whether the option was set.
//
//  last_set_sid
// Last set script id (if any).
//
//  last_set_linenr
// line number where option was set.
//
//  last_set_chan
// Channel where option was set (0 for local).
//
//  scope
// One of "global", "win", or "buf".
//
//  global_local
// Whether win or buf option has a global value.
//
//  commalist
// List of comma separated values.
//
//  flaglist
// List of single char flags.
func AllOptionsInfo() (opinfo OptionInfo) {
	name(nvim_get_all_options_info)
	returnPtr()
}

// OptionInfo gets the option information for one option.
//
// Resulting map has keys:
//
//  name
// Name of the option (like "filetype").
//
//  shortname
// Shortened name of the option (like "ft").
//
//  type
// type of option ("string", "number" or "boolean").
//
//  default
// The default value for the option.
//
//  was_set
// Whether the option was set.
//
//  last_set_sid
// Last set script id (if any).
//
//  last_set_linenr
// line number where option was set.
//
//  last_set_chan
// Channel where option was set (0 for local).
//
//  scope
// One of "global", "win", or "buf".
//
//  global_local
// Whether win or buf option has a global value.
//
//  commalist
// List of comma separated values.
//
//  flaglist
// List of single char flags.
func OptionInfo(name string) (opinfo OptionInfo) {
	name(nvim_get_option_info)
	returnPtr()
}

// SetOption sets an option value.
func SetOption(name string, value interface{}) {
	name(nvim_set_option)
}

// Echo echo a message.
//
// chunks is a list of [text, hl_group] arrays, each representing a
// text chunk with specified highlight. hl_group element can be omitted for no highlight.
//
// If history is true, add to "message-history".
//
// opts is optional parameters. Reserved for future use.
func Echo(chunks []TextChunk, history bool, opts map[string]interface{}) {
	name(nvim_echo)
}

// WriteOut writes a message to the Vim output buffer.
//
// Does not append "\n", the message is buffered (won't display) until a linefeed is written.
func WriteOut(str string) {
	name(nvim_out_write)
}

// WriteErr writes a message to the Vim error buffer.
//
// Does not append "\n", the message is buffered (won't display) until a linefeed is written.
func WriteErr(str string) {
	name(nvim_err_write)
}

// WritelnErr writes a message to the Vim error buffer.
//
// Appends "\n", so the buffer is flushed and displayed.
func WritelnErr(str string) {
	name(nvim_err_writeln)
}

// Buffers gets the current list of buffer handles.
func Buffers() (buffers []Buffer) {
	name(nvim_list_bufs)
}

// CurrentBuffer gets the current buffer.
func CurrentBuffer() (buffer Buffer) {
	name(nvim_get_current_buf)
}

// SetCurrentBuffer sets the current buffer.
func SetCurrentBuffer(buffer Buffer) {
	name(nvim_set_current_buf)
}

// Windows gets the current list of window handles.
func Windows() (windows []Window) {
	name(nvim_list_wins)
}

// CurrentWindow gets the current window.
func CurrentWindow() (window Window) {
	name(nvim_get_current_win)
}

// SetCurrentWindow sets the current window.
func SetCurrentWindow(window Window) {
	name(nvim_set_current_win)
}

// CreateBuffer greates a new, empty, unnamed buffer.
//
// The listed arg sets buflisted buffer opttion.
//
// The scratch arg creates a "throwaway" "scratch-buffer" for temporary work (always "nomodified").
// Also sets "nomodeline" on the buffer.
func CreateBuffer(listed, scratch bool) (buffer Buffer) {
	name(nvim_create_buf)
}

// OpenTerm opens a terminal instance in a buffer.
//
// By default (and currently the only option) the terminal will not be
// connected to an external process. Instead, input send on the channel
// will be echoed directly by the terminal. This is useful to disply
// ANSI terminal sequences returned as part of a rpc message, or similar.
//
// Note that to directly initiate the terminal using the right size, display the
// buffer in a configured window before calling this. For instance, for a
// floating display, first create an empty buffer using CreateBuffer,
// then display it using OpenWindow, and then call this function.
// Then "nvim_chan_send" cal be called immediately to process sequences
// in a virtual terminal having the intended size.
//
// The buffer arg is the buffer to use (expected to be empty).
//
// The opts arg is optional parameters. Reserved for future use.
func OpenTerm(buffer Buffer, opts map[string]interface{}) (channel int) {
	name(nvim_open_term)
}

// OpenWindow open a new window.
//
// Currently this is used to open floating and external windows.
// Floats are windows that are drawn above the split layout, at some anchor
// position in some other window.
// Floats can be drawn internally or by external GUI with the |ui-multigrid| extension.
// External windows are only supported with multigrid GUIs, and are displayed as separate top-level windows.
//
// For a general overview of floats, see
//  :help api-floatwin
//
// Exactly one of `external` and `relative` must be specified.
// The `width` and `height` of the new window must be specified.
//
// With relative=editor (row=0,col=0) refers to the top-left corner of the
// screen-grid and (row=Lines-1,col=Columns-1) refers to the bottom-right
// corner.
// Fractional values are allowed, but the builtin implementation
// (used by non-multigrid UIs) will always round down to nearest integer.
//
// Out-of-bounds values, and configurations that make the float not fit inside
// the main editor, are allowed.
//
// The builtin implementation truncates values so floats are fully within the main screen grid.
// External GUIs could let floats hover outside of the main window like a tooltip, but
// this should not be used to specify arbitrary WM screen positions.
func OpenWindow(buffer Buffer, enter bool, config *WindowConfig) (window Window) {
	name(nvim_open_win)
}

// Tabpages gets the current list of tabpage handles.
func Tabpages() (tabpages []Tabpage) {
	name(nvim_list_tabpages)
}

// CurrentTabpage gets the current tabpage.
func CurrentTabpage() (tabpage Tabpage) {
	name(nvim_get_current_tabpage)
}

// SetCurrentTabpage sets the current tabpage.
func SetCurrentTabpage(tabpage Tabpage) {
	name(nvim_set_current_tabpage)
}

// CreateNamespace creates a new namespace, or gets an existing one.
//
// Namespaces are used for buffer highlights and virtual text, see
// AddBufferHighlight and SetBufferVirtualText.
//
// Namespaces can be named or anonymous.
// If "name" matches an existing namespace, the associated id is returned.
// If "name" is an empty string a new, anonymous namespace is created.
//
// The returns the namespace ID.
func CreateNamespace(name string) (nsID int) {
	name(nvim_create_namespace)
}

// Namespaces gets existing, non-anonymous namespaces.
//
// The return dict that maps from names to namespace ids.
func Namespaces() (namespaces map[string]int) {
	name(nvim_get_namespaces)
}

// Paste pastes at cursor, in any mode.
//
// Invokes the `vim.paste` handler, which handles each mode appropriately.
// Sets redo/undo. Faster than Input(). Lines break at LF ("\n").
//
// Errors (`nomodifiable`, `vim.paste()` `failure` ...) are reflected in `err`
// but do not affect the return value (which is strictly decided by `vim.paste()`).
//
// On error, subsequent calls are ignored ("drained") until the next paste is initiated (phase 1 or -1).
//
//  data
// multiline input. May be binary (containing NUL bytes).
//
//  crlf
// also break lines at CR and CRLF.
//
//  phase
// -1 is paste in a single call (i.e. without streaming).
//
// To `stream` a paste, call Paste sequentially with these `phase` args:
//  1
// starts the paste (exactly once)
//  2
// continues the paste (zero or more times)
//  3
// ends the paste (exactly once)
//
// The returned  boolean state is:
//
//  true
// Client may continue pasting.
//
//  false
// Client must cancel the paste.
func Paste(data string, crlf bool, phase int) (state bool) {
	name(nvim_paste)
}

// Put puts text at cursor, in any mode.
//
// Compare :put and p which are always linewise.
//
// lines is readfile() style list of lines.
//
// typ is edit behavior: any getregtype() result, or:
//   "b"
//  blockwise-visual mode (may include width, e.g. "b3")
//   "c"
//  characterwise mode
//   "l"
//  linewise mode
//   ""
// guess by contents, see setreg().
//
// After is insert after cursor (like `p`), or before (like `P`).
//
// follow arg is place cursor at end of inserted text.
func Put(lines []string, typ string, after, follow bool) {
	name(nvim_put)
}

// Subscribe subscribes to a Nvim event.
func Subscribe(event string) {
	name(nvim_subscribe)
}

// Unsubscribe unsubscribes to a Nvim event.
func Unsubscribe(event string) {
	name(nvim_unsubscribe)
}

// ColorByName returns the 24-bit RGB value of a ColorMap color name or `#rrggbb` hexadecimal string.
func ColorByName(name string) (color int) {
	name(nvim_get_color_by_name)
}

// ColorMap returns a map of color names and RGB values.
//
// Keys are color names (e.g. "Aqua") and values are 24-bit RGB color values (e.g. 65535).
//
// The returns map is color names and RGB values.
func ColorMap() (colorMap map[string]int) {
	name(nvim_get_color_map)
}

// Context gets a map of the current editor state.
// This API still under development.
//
// The opts arg is optional parameters.
// Key is `types`.
//
// List of context-types to gather, or empty for `all` context.
//
//  regs
//  jumps
//  bufs
//  gvars
//  funcs
//  sfuncs
func Context(opts map[string][]string) (contexts map[string]interface{}) {
	name(nvim_get_context)
}

// LoadContext sets the current editor state from the given context map.
func LoadContext(dict map[string]interface{}) (contextMap interface{}) {
	name(nvim_load_context)
}

// Mode gets the current mode.
func Mode() (mode Mode) {
	name(nvim_get_mode)
	returnPtr()
}

// KeyMap gets a list of global (non-buffer-local) mapping definitions.
//
// The mode arg is the mode short-name, like `n`, `i`, `v` or etc.
func KeyMap(mode string) (maps []*Mapping) {
	name(nvim_get_keymap)
}

// SetKeyMap sets a global mapping for the given mode.
//
// To set a buffer-local mapping, use SetBufferKeyMap().
//
// Unlike :map, leading/trailing whitespace is accepted as part of the {lhs}
// or {rhs}.
// Empty {rhs} is <Nop>. keycodes are replaced as usual.
//
//  mode
// mode short-name (map command prefix: "n", "i", "v", "x", …) or "!" for :map!, or empty string for :map.
//
//  lhs
// Left-hand-side {lhs} of the mapping.
//
//  rhs
// Right-hand-side {rhs} of the mapping.
//
//  opts
// Optional parameters map. Accepts all ":map-arguments" as keys excluding "buffer" but including "noremap".
// Values are Booleans. Unknown key is an error.
func SetKeyMap(mode, lhs, rhs string, opts map[string]bool) {
	name(nvim_set_keymap)
}

// DeleteKeyMap unmaps a global mapping for the given mode.
//
// To unmap a buffer-local mapping, use DeleteBufferKeyMap().
//
// See:
//  :help nvim_set_keymap()
func DeleteKeyMap(mode, lhs string) {
	name(nvim_del_keymap)
}

// Commands gets a map of global (non-buffer-local) Ex commands.
// Currently only user-commands are supported, not builtin Ex commands.
//
// opts is optional parameters. Currently only supports:
//  {"builtin":false}
func Commands(opts map[string]interface{}) (commands map[string]*Command) {
	name(nvim_get_commands)
}

// APIInfo returns a 2-tuple (Array), where item 0 is the current channel id and item
// 1 is the "pi-metadata|" map (Dictionary).
//
// Returns 2-tuple [{channel-id}, {api-metadata}].
func APIInfo() (apiInfo []interface{}) {
	name(nvim_get_api_info)
}

// SetClientInfo self-identifies the client.
//
// The client/plugin/application should call this after connecting, to provide
// hints about its identity and purpose, for debugging and orchestration.
//
// Can be called more than once; the caller should merge old info if
// appropriate. Example: library first identifies the channel, then a plugin
// using that library later identifies itself.
func SetClientInfo(name string, version *ClientVersion, typ string, methods map[string]*ClientMethod, attributes ClientAttributes) {
	name(nvim_set_client_info)
}

// ChannelInfo get information about a channel.
//
// Rreturns Dictionary describing a channel, with these keys:
//
//  "stream"
// The stream underlying the channel. value are:
//  "stdio"
// stdin and stdout of this Nvim instance.
//  "stderr"
// stderr of this Nvim instance.
//  "socket"
// TCP/IP socket or named pipe.
//  "job"
// job with communication over its stdio.
//
//  "mode"
// How data received on the channel is interpreted. value are:
//  "bytes"
// send and receive raw bytes.
//  "terminal"
// A terminal instance interprets ASCII sequences.
//  "rpc"
// RPC communication on the channel is active.
//
//  "pty"
// Name of pseudoterminal, if one is used (optional).
// On a POSIX system, this will be a device path like /dev/pts/1.
// Even if the name is unknown, the key will still be present to indicate a pty is used.
// This is currently the case when using winpty on windows.
//
//  "buffer"
// Buffer with connected |terminal| instance (optional).
//
//  "client"
// Information about the client on the other end of the RPC channel, if it has added it using SetClientInfo() (optional).
func ChannelInfo(channelID int) (channel Channel) {
	name(nvim_get_chan_info)
	returnPtr()
}

// Channels get information about all open channels.
func Channels() (channels []*Channel) {
	name(nvim_list_chans)
}

// ParseExpression parse a VimL expression.
func ParseExpression(expr, flags string, highlight bool) (expression map[string]interface{}) {
	name(nvim_parse_expression)
}

// UIs gets a list of dictionaries representing attached UIs.
func UIs() (uis []*UI) {
	name(nvim_list_uis)
}

// ProcChildren gets the immediate children of process `pid`.
func ProcChildren(pid int) (processes []*Process) {
	name(nvim_get_proc_children)
}

// Proc gets info describing process `pid`.
func Proc(pid int) (process Process) {
	name(nvim_get_proc)
}

// SelectPopupmenuItem selects an item in the completion popupmenu.
//
// If |ins-completion| is not active this API call is silently ignored.
// Useful for an external UI using |ui-popupmenu| to control the popupmenu
// with the mouse. Can also be used in a mapping; use <cmd> |:map-cmd| to
// ensure the mapping doesn't end completion mode.
//
// The `opts` optional parameters. Reserved for future use.
func SelectPopupmenuItem(item int, insert, finish bool, opts map[string]interface{}) {
	name(nvim_select_popupmenu_item)
}

// TODO(zchee): nvim_set_decoration_provider

// buffer.c

// BufferLineCount gets the buffer line count.
//
// The buffer arg is specific Buffer, or 0 for current buffer.
//
// The returns line count, or 0 for unloaded buffer.
func BufferLineCount(buffer Buffer) (count int) {
	name(nvim_buf_line_count)
}

// AttachBuffer activates buffer-update events on a channel.
//
// The buffer is specific Buffer, or 0 for current buffer.
//
// If sendBuffer is true, initial notification should contain the whole buffer.
// If false, the first notification will be a `nvim_buf_lines_event`.
// Otherwise, the first notification will be a `nvim_buf_changedtick_event`
//
// Returns whether the updates couldn't be enabled because the buffer isn't loaded or opts contained an invalid key.
func AttachBuffer(buffer Buffer, sendBuffer bool, opts map[string]interface{}) (attached bool) {
	name(nvim_buf_attach)
}

// DetachBuffer deactivate updates from this buffer to the current channel.
//
// Returns whether the updates couldn't be disabled because the buffer isn't loaded.
func DetachBuffer(buffer Buffer) (detached bool) {
	name(nvim_buf_detach)
}

// BufferLines gets a line-range from the buffer.
//
// Indexing is zero-based, end-exclusive.
// Negative indices are interpreted as length+1+index: -1 refers to the index past the end.
// So to get the last element use start=-2 and end=-1.
//
// Out-of-bounds indices are clamped to the nearest valid value, unless strictIndexing is set.
func BufferLines(buffer Buffer, start, end int, strictIndexing bool) (lines [][]byte) {
	name(nvim_buf_get_lines)
}

// SetBufferLines sets or replaces a line-range in the buffer.
//
// Indexing is zero-based, end-exclusive.
// Negative indices are interpreted as length+1+index: -1 refers to the index past the end.
// So to change or delete the last element use start=-2 and end=-1.
//
// To insert lines at a given index, set start and end args to the same index.
//
// To delete a range of lines, set replacement arg to an empty array.
//
// Out-of-bounds indices are clamped to the nearest valid value, unless
// strict_indexing arg is set to true.
func SetBufferLines(buffer Buffer, start, end int, strictIndexing bool, replacement [][]byte) {
	name(nvim_buf_set_lines)
}

// SetBufferText sets or replaces a range in the buffer.
//
// This is recommended over SetBufferLines when only modifying parts of a
// line, as extmarks will be preserved on non-modified parts of the touched
// lines.
//
// Indexing is zero-based and end-exclusive.
//
// To insert text at a given index, set startRow and endRow args ranges to the same index.
//
// To delete a range, set replacement arg to an array containing an empty string, or simply an empty array.
//
// Prefer SetBufferLines when adding or deleting entire lines only.
func SetBufferText(buffer Buffer, startRow, startCol, endRow, endCol int, replacement [][]byte) {
	name(nvim_buf_set_text)
}

// BufferOffset returns the byte offset for a line.
//
// Line 1 (index=0) has offset 0. UTF-8 bytes are counted. EOL is one byte.
// 'fileformat' and 'fileencoding' are ignored.
//
// The line index just after the last line gives the total byte-count of the buffer.
// A final EOL byte is counted if it would be written, see ':help eol'.
//
// Unlike `line2byte` vim function, throws error for out-of-bounds indexing.
//
// If Buffer is unloaded buffer, returns -1.
func BufferOffset(buffer Buffer, index int) (offset int) {
	name(nvim_buf_get_offset)
}

// BufferVar gets a buffer-scoped (b:) variable.
func BufferVar(buffer Buffer, name string) (value interface{}) {
	name(nvim_buf_get_var)
}

// BufferChangedTick gets a changed tick of a buffer.
func BufferChangedTick(buffer Buffer) (changedtick int) {
	name(nvim_buf_get_changedtick)
}

// BufferKeymap gets a list of buffer-local mapping definitions.
//
// The mode short-name ("n", "i", "v", ...).
func BufferKeyMap(buffer Buffer, mode string) []*Mapping {
	name(nvim_buf_get_keymap)
}

// SetBufferKeyMap sets a buffer-local mapping for the given mode.
//
// See:
//  :help nvim_set_keymap()
func SetBufferKeyMap(buffer Buffer, mode, lhs, rhs string, opts map[string]bool) {
	name(nvim_buf_set_keymap)
}

// DeleteBufferKeyMap unmaps a buffer-local mapping for the given mode.
//
// See:
//  :help nvim_del_keymap()
func DeleteBufferKeyMap(buffer Buffer, mode, lhs string) {
	name(nvim_buf_del_keymap)
}

// BufferCommands gets a map of buffer-local user-commands.
//
// opts is optional parameters. Currently not used.
func BufferCommands(buffer Buffer, opts map[string]interface{}) map[string]*Command {
	name(nvim_buf_get_commands)
}

// SetBufferVar sets a buffer-scoped (b:) variable.
func SetBufferVar(buffer Buffer, name string, value interface{}) {
	name(nvim_buf_set_var)
}

// DeleteBufferVar removes a buffer-scoped (b:) variable.
func DeleteBufferVar(buffer Buffer, name string) {
	name(nvim_buf_del_var)
}

// BufferOption gets a buffer option value.
func BufferOption(buffer Buffer, name string) (value interface{}) {
	name(nvim_buf_get_option)
}

// SetBufferOption sets a buffer option value.
//
// Passing nil as value arg to deletes the option (only works if there's a global fallback).
func SetBufferOption(buffer Buffer, name string, value interface{}) {
	name(nvim_buf_set_option)
}

// BufferNumber gets a buffer's number.
//
// Deprecated: Use int(buffer) to get the buffer's number as an integer.
func BufferNumber(buffer Buffer) (number int) {
	name(nvim_buf_get_number)
	deprecatedSince(2)
}

// BufferName gets the full file name for the buffer.
func BufferName(buffer Buffer) (name string) {
	name(nvim_buf_get_name)
}

// SetBufferName sets the full file name for a buffer.
func SetBufferName(buffer Buffer, name string) {
	name(nvim_buf_set_name)
}

// IsBufferLoaded checks if a buffer is valid and loaded.
//
// See |help api-buffer| for more info about unloaded buffers.
func IsBufferLoaded(buffer Buffer) (loaded bool) {
	name(nvim_buf_is_loaded)
}

// DeleteBuffer deletes the buffer.
// See
//  help :bwipeout
//
// The opts args is optional parameters.
//
//  force
// Force deletion and ignore unsaved changes. bool type.
//
//  unload
// Unloaded only, do not delete. See |help :bunload|. bool type.
func DeleteBuffer(buffer Buffer, opts map[string]bool) {
	name(nvim_buf_delete)
}

// IsBufferValid returns whether the buffer is valid.
//
// Note: Even if a buffer is valid it may have been unloaded.
// See |help api-buffer| for more info about unloaded buffers.
func IsBufferValid(buffer Buffer) (valied bool) {
	name(nvim_buf_is_valid)
}

// BufferMark return a tuple (row,col) representing the position of the named mark.
//
// Marks are (1,0)-indexed.
func BufferMark(buffer Buffer, name string) (pos [2]int) {
	name(nvim_buf_get_mark)
}

// BufferExtmarkByID beturns position for a given extmark id.
//
// The opts arg is optional parameters.
//
//  limit
// Maximum number of marks to return. int type.
//
//  details
// Whether to include the details dict. bool type.
func BufferExtmarkByID(buffer Buffer, nsID, id int, opt map[string]interface{}) (pos []int) {
	name(nvim_buf_get_extmark_by_id)
}

// BufferExtmarks gets extmarks in "traversal order" from a |charwise| region defined by
// buffer positions (inclusive, 0-indexed).
//
// Region can be given as (row,col) tuples, or valid extmark ids (whose
// positions define the bounds).
// 0 and -1 are understood as (0,0) and (-1,-1) respectively, thus the following are equivalent:
//
//   BufferExtmarks(0, myNS, 0, -1, {})
//   BufferExtmarks(0, myNS, [0,0], [-1,-1], {})
//
// If end arg is less than start arg, traversal works backwards.
// It useful with limit arg, to get the first marks prior to a given position.
//
// The start and end args is start or end of range, given as (row, col), or
// valid extmark id whose position defines the bound.
//
// The opts arg is optional parameters.
//
//  limit
// Maximum number of marks to return. int type.
//
//  details
// Whether to include the details dict. bool type.
func BufferExtmarks(buffer Buffer, nsID int, start, end interface{}, opt map[string]interface{}) (marks []ExtMark) {
	name(nvim_buf_get_extmarks)
}

// SetBufferExtmark creates or updates an extmark.
//
// To create a new extmark, pass id=0. The extmark id will be returned.
// To move an existing mark, pass its id.
//
// It is also allowed to create a new mark by passing in a previously unused
// id, but the caller must then keep track of existing and unused ids itself.
// (Useful over RPC, to avoid waiting for the return value.)
//
// Using the optional arguments, it is possible to use this to highlight
// a range of text, and also to associate virtual text to the mark.
//
// The opts arg is optional parameters.
//
//  id
// ID of the extmark to edit.
//
//  end_line
// Ending line of the mark, 0-based inclusive.
//
//  end_col
// Ending col of the mark, 0-based inclusive.
//
//  hl_group
// Name of the highlight group used to highlight this mark.
//
//  virt_text
// Virtual text to link to this mark.
//
//  virt_text_pos
// Positioning of virtual text.
// Possible values:
//  "eol"
// right after eol character (default)
//  "overlay"
// display over the specified column, without shifting the underlying text.
//
//  virt_text_hide
// Hide the virtual text when the background text is selected or hidden due to horizontal scroll 'nowrap'.
//
//  hl_mode
// Control how highlights are combined with the highlights of the text. Currently only affects
// virt_text highlights, but might affect "hl_group" in later versions.
// Possible values:
//  "replace"
// only show the virt_text color. This is the default.
//  "combine"
// combine with background text color
//  "blend"
// blend with background text color.
//
//  ephemeral
// For use with "nvim_set_decoration_provider" callbacks. The mark will only be used for the current redraw cycle,
// and not be permantently stored in the buffer.
//
//  right_gravity
// Boolean that indicates the direction the extmark will be shifted in when new text is
// inserted (true for right, false for left).  defaults to true.
//
//  end_right_gravity
// Boolean that indicates the direction the extmark end position (if it exists) will be
// shifted in when new text is inserted (true for right, false for left). Defaults to false.
func SetBufferExtmark(buffer Buffer, nsID, line, col int, opts map[string]interface{}) (id int) {
	name(nvim_buf_set_extmark)
}

// DeleteBufferExtmark removes an extmark.
//
// THe returns whether the extmark was found.
func DeleteBufferExtmark(buffer Buffer, nsID, extmarkID int) (deleted bool) {
	name(nvim_buf_del_extmark)
}

// AddBufferHighlight adds a highlight to buffer.
//
// IT useful for plugins that dynamically generate highlights to a buffer like a semantic highlighter or linter.
//
// The function adds a single highlight to a buffer.
// Unlike |matchaddpos()| vim function, highlights follow changes to line numbering as lines are
// inserted/removed above the highlighted line, like signs and marks do.
//
// Namespaces are used for batch deletion/updating of a set of highlights.
// To create a namespace, use CreateNamespace which returns a namespace id.
// Pass it in to this function as nsID to add highlights to the namespace.
// All highlights in the same namespace can then be cleared with single call to ClearBufferNamespace.
// If the highlight never will be deleted by an API call, pass nsID = -1.
//
// As a shorthand, `nsID = 0` can be used to create a new namespace for the
// highlight, the allocated id is then returned.
//
// If hlGroup arg is the empty string, no highlight is added, but a new `nsID` is still returned.
// This is supported for backwards compatibility, new code should use CreateNamespaceto create a new empty namespace.
func AddBufferHighlight(buffer Buffer, srcID int, hlGroup string, line, startCol, endCol int) (id int) {
	name(nvim_buf_add_highlight)
}

// ClearBufferNamespace clears namespaced objects (highlights, extmarks, virtual text) from a region.
// Lines are 0-indexed.
//
// To clear the namespace in the entire buffer, specify line_start=0 and line_end=-1.
func ClearBufferNamespace(buffer Buffer, nsID, lineStart, lineEnd int) {
	name(nvim_buf_clear_namespace)
}

// ClearBufferHighlight clears highlights from a given source group and a range
// of lines.
//
// To clear a source group in the entire buffer, pass in 1 and -1 to startLine
// and endLine respectively.
//
// The lineStart and lineEnd parameters specify the range of lines to clear.
// The end of range is exclusive. Specify -1 to clear to the end of the file.
//
// Deprecated: Use ClearBufferNamespace() instead.
func ClearBufferHighlight(buffer Buffer, srcID, startLine, endLine int) {
	name(nvim_buf_clear_highlight)
	deprecatedSince(7)
}

// SetBufferVirtualText set the virtual text (annotation) for a buffer line.
//
// By default (and currently the only option), the text will be placed after
// the buffer text.
//
// Virtual text will never cause reflow, rather virtual text will be truncated at the end of the screen line.
// The virtual text will begin one cell (|lcs-eol| or space) after the ordinary text.
//
// Namespaces are used to support batch deletion/updating of virtual text.
// To create a namespace, use CreateNamespace. Virtual text is cleared using ClearBufferNamespace.
//
// The same nsID can be used for both virtual text and highlights added by AddBufferHighlight,
// both can then be cleared with a single call to ClearBufferNamespace.
// If the virtual text never will be cleared by an API call, pass `nsID = -1`.
//
// As a shorthand, `nsID = 0` can be used to create a new namespace for the
// virtual text, the allocated id is then returned.
//
// The opts arg is reserved for future use.
func SetBufferVirtualText(buffer Buffer, nsID, line int, chunks []TextChunk, opts map[string]interface{}) (id int) {
	name(nvim_buf_set_virtual_text)
}

// window.c

// WindowBuffer returns the current buffer in a window.
func WindowBuffer(window Window) (buffer Buffer) {
	name(nvim_win_get_buf)
}

// SetBufferToWindow sets the current buffer in a window, without side-effects.
func SetBufferToWindow(window Window, buffer Buffer) {
	name(nvim_win_set_buf)
}

// WindowCursor returns the cursor position in the window.
func WindowCursor(window Window) (pos [2]int) {
	name(nvim_win_get_cursor)
}

// SetWindowCursor sets the cursor position in the window to the given position.
func SetWindowCursor(window Window, pos [2]int) {
	name(nvim_win_set_cursor)
}

// WindowHeight returns the window height.
func WindowHeight(window Window) (height int) {
	name(nvim_win_get_height)
}

// SetWindowHeight sets the window height.
func SetWindowHeight(window Window, height int) {
	name(nvim_win_set_height)
}

// WindowWidth returns the window width.
func WindowWidth(window Window) (width int) {
	name(nvim_win_get_width)
}

// SetWindowWidth sets the window width.
func SetWindowWidth(window Window, width int) {
	name(nvim_win_set_width)
}

// WindowVar gets a window-scoped (w:) variable.
func WindowVar(window Window, name string) (value interface{}) {
	name(nvim_win_get_var)
}

// SetWindowVar sets a window-scoped (w:) variable.
func SetWindowVar(window Window, name string, value interface{}) {
	name(nvim_win_set_var)
}

// DeleteWindowVar removes a window-scoped (w:) variable.
func DeleteWindowVar(window Window, name string) {
	name(nvim_win_del_var)
}

// WindowOption gets a window option.
func WindowOption(window Window, name string) (value interface{}) {
	name(nvim_win_get_option)
}

// SetWindowOption sets a window option.
func SetWindowOption(window Window, name string, value interface{}) {
	name(nvim_win_set_option)
}

// WindowPosition gets the window position in display cells. First position is zero.
func WindowPosition(window Window) (pos [2]int) {
	name(nvim_win_get_position)
}

// WindowTabpage gets the tab page that contains the window.
func WindowTabpage(window Window) (tabpage Tabpage) {
	name(nvim_win_get_tabpage)
}

// WindowNumber gets the window number from the window handle.
func WindowNumber(window Window) (number int) {
	name(nvim_win_get_number)
}

// IsWindowValid returns true if the window is valid.
func IsWindowValid(window Window) (valid bool) {
	name(nvim_win_is_valid)
}

// SetWindowConfig configure window position. Currently this is only used to configure
// floating and external windows (including changing a split window to these types).
//
// See documentation at OpenWindow, for the meaning of parameters.
//
// When reconfiguring a floating window, absent option keys will not be
// changed.
// The following restriction are apply must be reconfigured together. Only changing a subset of these is an error.
//  row
//  col
//  relative
func SetWindowConfig(window Window, config *WindowConfig) {
	name(nvim_win_set_config)
}

// WindowConfig return window configuration.
//
// Return a dictionary containing the same config that can be given to OpenWindow.
//
// The `relative` will be an empty string for normal windows.
func WindowConfig(window Window) (config WindowConfig) {
	name(nvim_win_get_config)
	returnPtr()
}

// HideWindow closes the window and hide the buffer it contains (like ":hide" with a
// windowID).
//
// Like ":hide" the buffer becomes hidden unless another window is editing it,
// or "bufhidden" is "unload", "delete" or "wipe" as opposed to ":close" or
// CloseWindow, which will close the buffer.
func HideWindow(window Window) {
	name(nvim_win_hide)
}

// CloseWindow close a window.
//
// This is equivalent to |:close| with count except that it takes a window id.
func CloseWindow(window Window, force bool) {
	name(nvim_win_close)
}

// tabpage.c

// TabpageWindows returns the windows in a tabpage.
func TabpageWindows(tabpage Tabpage) (windows []Window) {
	name(nvim_tabpage_list_wins)
}

// TabpageVar gets a tab-scoped (t:) variable.
func TabpageVar(tabpage Tabpage, name string) (value interface{}) {
	name(nvim_tabpage_get_var)
}

// SetTabpageVar sets a tab-scoped (t:) variable.
func SetTabpageVar(tabpage Tabpage, name string, value interface{}) {
	name(nvim_tabpage_set_var)
}

// DeleteTabpageVar removes a tab-scoped (t:) variable.
func DeleteTabpageVar(tabpage Tabpage, name string) {
	name(nvim_tabpage_del_var)
}

// TabpageWindow gets the current window in a tab page.
func TabpageWindow(tabpage Tabpage) Window {
	name(nvim_tabpage_get_win)
}

// TabpageNumber gets the tabpage number from the tabpage handle.
func TabpageNumber(tabpage Tabpage) (number int) {
	name(nvim_tabpage_get_number)
}

// IsTabpageValid checks if a tab page is valid.
func IsTabpageValid(tabpage Tabpage) (valid bool) {
	name(nvim_tabpage_is_valid)
}

// ui.c

// AttachUI registers the client as a remote UI. After this method is called,
// the client will receive redraw notifications.
//
//  :help rpc-remote-ui
//
// The redraw notification method has variadic arguments. Register a handler
// for the method like this:
//
//  v.RegisterHandler("redraw", func(updates ...[]interface{}) {
//      for _, update := range updates {
//          // handle update
//      }
//  })
func AttachUI(width, height int, options map[string]interface{}) {
	name(nvim_ui_attach)
}

// DetachUI unregisters the client as a remote UI.
func DetachUI() {
	name(nvim_ui_detach)
}

// TryResizeUI notifies Nvim that the client window has resized. If possible,
// Nvim will send a redraw request to resize.
func TryResizeUI(width, height int) {
	name(nvim_ui_try_resize)
}

// SetUIOption sets a UI option.
func SetUIOption(name string, value interface{}) {
	name(nvim_ui_set_option)
}

// TryResizeUIGrid tell Nvim to resize a grid. Triggers a grid_resize event with the requested
// grid size or the maximum size if it exceeds size limits.
//
// On invalid grid handle, fails with error.
func TryResizeUIGrid(grid, width, height int) {
	name(nvim_ui_try_resize_grid)
}

// SetPumHeight tells Nvim the number of elements displaying in the popumenu, to decide
// <PageUp> and <PageDown> movement.
//
// height is popupmenu height, must be greater than zero.
func SetPumHeight(height int) {
	name(nvim_ui_pum_set_height)
}

// SetPumBounds tells Nvim the geometry of the popumenu, to align floating windows with an
// external popup menu.
//
// Note that this method is not to be confused with SetPumHeight,
// which sets the number of visible items in the popup menu, while this
// function sets the bounding box of the popup menu, including visual
// elements such as borders and sliders.
//
// Floats need not use the same font size, nor be anchored to exact grid corners, so one can set floating-point
// numbers to the popup menu geometry.
func SetPumBounds(width, height, row, col float64) {
	name(nvim_ui_pum_set_bounds)
}
