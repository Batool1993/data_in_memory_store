package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"data_storage/client"
)

// CLIArgs holds dynamic inputs from flags.
type CLIArgs struct {
	Action      string
	Key         string
	Value       string
	Values      []string
	TTLOverride time.Duration
	Interval    time.Duration
	Timeout     time.Duration
}

// ParseArgs defines and validates flags.
func ParseArgs(defaultTTL time.Duration) (*CLIArgs, error) {
	action := flag.String("action", "", "one of: set|get|del|lpush|rpop")
	key := flag.String("key", "", "key to operate on")
	value := flag.String("value", "", "value for set or single lpush")
	values := flag.String("values", "", "comma-separated values for lpush")
	ttl := flag.Duration("ttl", 0, "override TTL (e.g. 30s); omit to use default")
	interval := flag.Duration("interval", 0, "cleanup interval for background tasks (e.g. 30s)")
	timeout := flag.Duration("timeout", defaultTTL+5*time.Second, "request timeout")
	flag.Parse()

	if *action == "" {
		return nil, fmt.Errorf("--action is required")
	}
	if *key == "" {
		return nil, fmt.Errorf("--key is required")
	}

	var vals []string
	if *values != "" {
		vals = strings.Split(*values, ",")
	}

	return &CLIArgs{
		Action:      *action,
		Key:         *key,
		Value:       *value,
		Values:      vals,
		TTLOverride: *ttl,
		Interval:    *interval,
		Timeout:     *timeout,
	}, nil
}

// CLI ties flag parsing to the StoreClient interface.
type CLI struct {
	store      client.StoreClient
	defaultTTL time.Duration
}

// NewCLI injects the StoreClient and default TTL.
func NewCLI(store client.StoreClient, defaultTTL time.Duration) *CLI {
	return &CLI{store: store, defaultTTL: defaultTTL}
}

// commandFunc is the signature of each command handler.
type commandFunc func(ctx context.Context, args *CLIArgs) error

// Run dispatches to the appropriate commandFunc.
func (cli *CLI) Run(args *CLIArgs) error {
	cmds := map[string]commandFunc{
		"set":   cli.runSet,
		"get":   cli.runGet,
		"del":   cli.runDelete,
		"lpush": cli.runLPush,
		"rpop":  cli.runRPop,
	}

	fn, ok := cmds[args.Action]
	if !ok {
		return fmt.Errorf("unknown action %q; use set|get|del|lpush|rpop", args.Action)
	}

	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	return fn(ctx, args)
}

func (cli *CLI) runSet(ctx context.Context, args *CLIArgs) error {
	if args.Value == "" {
		return fmt.Errorf("--value is required for set")
	}
	ttl := args.TTLOverride
	if ttl == 0 {
		ttl = cli.defaultTTL
	}
	return cli.store.SetString(ctx, args.Key, args.Value, ttl)
}

func (cli *CLI) runGet(ctx context.Context, args *CLIArgs) error {
	v, err := cli.store.GetString(ctx, args.Key)
	if err != nil {
		return err
	}
	fmt.Println(v)
	return nil
}

func (cli *CLI) runDelete(ctx context.Context, args *CLIArgs) error {
	return cli.store.DeleteString(ctx, args.Key)
}

func (cli *CLI) runLPush(ctx context.Context, args *CLIArgs) error {
	if len(args.Values) == 0 {
		return fmt.Errorf("--values is required for lpush")
	}
	return cli.store.LPush(ctx, args.Key, args.Values...)
}

func (cli *CLI) runRPop(ctx context.Context, args *CLIArgs) error {
	v, err := cli.store.RPop(ctx, args.Key)
	if err != nil {
		return err
	}
	fmt.Println(v)
	return nil
}
