package cli_test

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"data_storage/client/cli"
)

// stubStoreClient implements client.StoreClient for testing.
type stubStoreClient struct {
	setCalled   bool
	setKey      string
	setValue    string
	setTTL      time.Duration
	getValue    string
	getCalled   bool
	delCalled   bool
	lpushCalled bool
	lpushKey    string
	lpushItems  []string
	rpopValue   string
	rpopCalled  bool
}

func (s *stubStoreClient) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	s.setCalled = true
	s.setKey = key
	s.setValue = value
	s.setTTL = ttl
	return nil
}

func (s *stubStoreClient) GetString(ctx context.Context, key string) (string, error) {
	s.getCalled = true
	return s.getValue, nil
}

func (s *stubStoreClient) DeleteString(ctx context.Context, key string) error {
	s.delCalled = true
	return nil
}

func (s *stubStoreClient) LPush(ctx context.Context, key string, items ...string) error {
	s.lpushCalled = true
	s.lpushKey = key
	s.lpushItems = append([]string(nil), items...)
	return nil
}

func (s *stubStoreClient) RPop(ctx context.Context, key string) (string, error) {
	s.rpopCalled = true
	return s.rpopValue, nil
}

func TestCLI_Run_SetGetDeleteLPushRPop(t *testing.T) {
	defaultTTL := 30 * time.Second
	stub := &stubStoreClient{getValue: "hello", rpopValue: "world"}
	app := cli.NewCLI(stub, defaultTTL)

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Helper to run a command and capture output
	run := func(args []string) string {
		// Reset flags and args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = append([]string{"cmd"}, args...)

		// Parse flags into CLIArgs
		iargs, err := cli.ParseArgs(defaultTTL)
		if err != nil {
			t.Fatalf("ParseArgs failed: %v", err)
		}

		// Run command
		err = app.Run(iargs)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		// Read output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)

		// Prepare for next capture
		r, w, _ = os.Pipe()
		os.Stdout = w

		return strings.TrimSpace(buf.String())
	}

	// 1) Test set
	setOutput := run([]string{"--action=set", "--key=foo", "--value=bar", "--ttl=10s"})
	if stub.setCalled != true {
		t.Error("expected SetString to be called")
	}
	if stub.setKey != "foo" || stub.setValue != "bar" || stub.setTTL != 10*time.Second {
		t.Errorf("SetString called with wrong args: %v %v %v", stub.setKey, stub.setValue, stub.setTTL)
	}
	if setOutput != "" {
		t.Errorf("expected no output for set, got %q", setOutput)
	}

	// 2) Test get
	getOutput := run([]string{"--action=get", "--key=foo"})
	if !stub.getCalled {
		t.Error("expected GetString to be called")
	}
	if getOutput != "hello" {
		t.Errorf("expected get to print 'hello', got %q", getOutput)
	}

	// 3) Test delete
	stub.delCalled = false
	delOutput := run([]string{"--action=del", "--key=foo"})
	if !stub.delCalled {
		t.Error("expected DeleteString to be called")
	}
	if delOutput != "" {
		t.Errorf("expected no output for delete, got %q", delOutput)
	}

	// 4) Test lpush
	stub.lpushCalled = false
	lpushOutput := run([]string{"--action=lpush", "--key=mylist", "--values=a,b"})
	if !stub.lpushCalled {
		t.Error("expected LPush to be called")
	}
	if stub.lpushKey != "mylist" || len(stub.lpushItems) != 2 || stub.lpushItems[0] != "a" {
		t.Errorf("LPush called with wrong args: %v %v", stub.lpushKey, stub.lpushItems)
	}
	if lpushOutput != "" {
		t.Errorf("expected no output for lpush, got %q", lpushOutput)
	}

	// 5) Test rpop
	rpopOutput := run([]string{"--action=rpop", "--key=mylist"})
	if !stub.rpopCalled {
		t.Error("expected RPop to be called")
	}
	if rpopOutput != "world" {
		t.Errorf("expected rpop to print 'world', got %q", rpopOutput)
	}

	// Restore stdout
	w.Close()
	os.Stdout = origStdout
}
