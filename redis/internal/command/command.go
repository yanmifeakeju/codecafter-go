package command

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/resp"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/store"
)

type Command struct {
	store *store.Store
}

func New(s *store.Store) *Command {
	if s == nil {
		s = store.New(nil)
	}

	return &Command{store: s}
}

func (c *Command) Run(req resp.Value) (resp.Value, error) {
	if req.Type != resp.Array || req.Null || len(req.Array) == 0 {
		return resp.ErrorValue("ERR expected array command"), nil
	}

	if req.Array[0].Type != resp.BulkString || req.Array[0].Null {
		return resp.ErrorValue("ERR invalid command"), nil
	}

	args := make([]string, 0, len(req.Array)-1)
	for _, item := range req.Array[1:] {
		if item.Type != resp.BulkString || item.Null {
			return resp.ErrorValue("ERR invalid command argument"), nil
		}

		args = append(args, item.Str)
	}

	name := strings.ToLower(req.Array[0].Str)

	switch name {
	case "ping":
		return c.ping(args...)
	case "echo":
		return c.echo(args...)
	case "set":
		return c.set(args...)
	case "get":
		return c.get(args...)

	// List commands
	case "lpush":
		return c.lpush(args...)
	case "rpush":
		return c.rpush(args...)
	case "lrange":
		return c.lrange(args...)
	case "llen":
		return c.llen(args...)
	case "lpop":
		return c.lpop(args...)
	case "blpop":
		return c.blpop(args...)
	default:
		return resp.ErrorValue("ERR unknown command"), nil
	}

}

func (c *Command) ping(args ...string) (resp.Value, error) {
	if len(args) > 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'ping' command"), nil
	}

	if len(args) == 1 {
		return resp.BulkStringValue(args[0]), nil
	}

	return resp.SimpleStringValue("PONG"), nil
}

func (c *Command) echo(args ...string) (resp.Value, error) {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'echo' command"), nil
	}

	return resp.BulkStringValue(args[0]), nil
}

func (c *Command) set(args ...string) (resp.Value, error) {
	var key string
	var value string
	var expiresAt time.Time
	switch len(args) {
	case 2:
		key = args[0]
		value = args[1]
	case 4:
		key = args[0]
		value = args[1]
		expType := strings.ToLower(args[2])
		expValue := args[3]

		if expType != "ex" && expType != "px" {
			return resp.ErrorValue("ERR syntax error"), nil
		}

		d, err := strconv.ParseInt(expValue, 10, 64)

		if err != nil {
			return resp.ErrorValue("ERR value is not an integer or out of range"), nil
		}

		if d <= 0 {
			return resp.ErrorValue("ERR invalid expire time in 'set' command"), nil
		}

		if expType == "ex" {
			expiresAt = time.Now().Add(time.Second * time.Duration(d))
		} else {
			expiresAt = time.Now().Add(time.Millisecond * time.Duration(d))
		}
	default:
		return resp.ErrorValue("ERR wrong number of arguments for 'set' command"), nil
	}

	c.store.SetString(key, value, expiresAt)
	return resp.SimpleStringValue("OK"), nil
}

func (c *Command) get(args ...string) (resp.Value, error) {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'get' command"), nil
	}

	key := args[0]
	val, ok, err := c.store.GetString(key)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}

		return resp.Value{}, err
	}

	if !ok {
		return resp.NullBulkString(), nil
	}

	return resp.BulkStringValue(val), nil
}

func (c *Command) rpush(args ...string) (resp.Value, error) {
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'rpush' command"), nil
	}

	key := args[0]
	items := args[1:]

	n, err := c.store.RPush(key, items...)

	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}

		return resp.Value{}, err
	}

	return resp.IntValue(int64(n)), nil
}

func (c *Command) lpush(args ...string) (resp.Value, error) {
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'lpush' command"), nil
	}

	key := args[0]
	items := args[1:]

	n, err := c.store.LPush(key, items...)

	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}

		return resp.Value{}, err
	}

	return resp.IntValue(int64(n)), nil
}

func (c *Command) lrange(args ...string) (resp.Value, error) {
	if len(args) != 3 {
		return resp.ErrorValue("ERR wrong number of arguments for 'lrange' command"), nil
	}

	key := args[0]

	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.ErrorValue("ERR value is not an integer or out of range"), nil
	}

	end, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorValue("ERR value is not an integer or out of range"), nil
	}

	items, err := c.store.LRange(key, int(start), int(end))
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}

		return resp.ArrayValue([]resp.Value{}), err
	}

	values := make([]resp.Value, 0, len(items))
	for _, item := range items {
		values = append(values, resp.BulkStringValue(item))
	}

	return resp.ArrayValue(values), nil
}

func (c *Command) llen(args ...string) (resp.Value, error) {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'llen' command"), nil
	}

	key := args[0]

	n, err := c.store.LLen(key)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}

		return resp.Value{}, err
	}

	return resp.IntValue(int64(n)), nil
}

func (c *Command) lpop(args ...string) (resp.Value, error) {
	switch len(args) {
	case 1:
		key := args[0]
		v, ok, err := c.store.LPop(key)

		if err != nil {
			if errors.Is(err, store.ErrWrongType) {
				return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
			}

			return resp.Value{}, err
		}

		if !ok {
			return resp.NullBulkString(), nil
		}

		return resp.BulkStringValue(v), nil

	case 2:
		key := args[0]
		count, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return resp.ErrorValue("ERR value is not an integer or out of range"), nil
		}

		if count < 0 {
			return resp.ErrorValue("ERR value is out of range, must be positive"), nil
		}

		items, err := c.store.LPopN(key, int(count))

		if err != nil {
			if errors.Is(err, store.ErrWrongType) {
				return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
			}

			return resp.Value{}, err
		}

		v := make([]resp.Value, len(items))

		for i := range items {
			v[i] = resp.BulkStringValue(items[i])
		}

		return resp.ArrayValue(v), nil
	default:
		return resp.ErrorValue("ERR wrong number of arguments for 'lpop' command"), nil
	}
}

func (c *Command) blpop(args ...string) (resp.Value, error) {
	if len(args) != 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'blpop' command"), nil
	}

	key := args[0]

	timeout, err := strconv.ParseFloat(args[1], 10)
	if err != nil {
		return resp.ErrorValue("ERR invalid timeout"), nil
	}

	if timeout < 0 {
		return resp.ErrorValue("ERR timeout is negative"), nil
	}

	value, ok, err := c.store.BLPop(key, time.Duration(timeout*float64(time.Second)))
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value"), nil
		}
		return resp.Value{}, err
	}

	if !ok {
		return resp.Value{Type: resp.Array, Null: true}, nil
	}

	return resp.ArrayValue([]resp.Value{
		resp.BulkStringValue(key),
		resp.BulkStringValue(value),
	}), nil
}
