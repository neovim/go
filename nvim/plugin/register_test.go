package plugin_test

import (
	"os"
	"strings"
	"testing"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

func newEmbeddedPlugin(t *testing.T) (*plugin.Plugin, func()) {
	env := []string{}
	if v := os.Getenv("VIM"); v != "" {
		env = append(env, "VIM="+v)
	}

	opts := []nvim.ChildProcessOption{
		nvim.ChildProcessCommand(nvim.BinaryName),
		nvim.ChildProcessArgs("-u", "NONE", "-n", "--embed"),
		nvim.ChildProcessEnv(env),
		nvim.ChildProcessLogf(t.Logf),
	}
	v, err := nvim.NewChildProcess(opts...)
	if err != nil {
		t.Fatal(err)
	}

	return plugin.New(v), func() {
		if err := v.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRegister(t *testing.T) {
	p, cleanup := newEmbeddedPlugin(t)
	defer cleanup()

	p.Handle("hello", func(s string) (string, error) {
		return "Hello, " + s, nil
	})
	p.HandleFunction(&plugin.FunctionOptions{Name: "Hello"}, func(args []string) (string, error) {
		return "Hello, " + strings.Join(args, " "), nil
	})

	if err := p.RegisterForTests(); err != nil {
		t.Fatal(err)
	}

	{
		result, err := p.Nvim.CommandOutput(":echo Hello('John', 'Doe')")
		if err != nil {
			t.Error(err)
		}
		expected := "Hello, John Doe"
		if result != expected {
			t.Errorf("Hello returned %q, want %q", result, expected)
		}
	}

	{
		cid := p.Nvim.ChannelID()

		var result string
		if err := p.Nvim.Call("rpcrequest", &result, cid, "hello", "world"); err != nil {
			t.Fatal(err)
		}

		expected := "Hello, world"
		if result != expected {
			t.Errorf("hello returned %q, want %q", result, expected)
		}
	}
}
