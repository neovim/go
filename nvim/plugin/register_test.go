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
	v, err := nvim.NewChildProcess(
		nvim.ChildProcessArgs("-u", "NONE", "-n", "--embed"),
		nvim.ChildProcessEnv(env),
		nvim.ChildProcessLogf(t.Logf))
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
