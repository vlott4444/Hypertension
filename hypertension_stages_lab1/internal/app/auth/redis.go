package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Redis calls are fail-closed: protected requests fail if Redis is unavailable.
func redisCommand(ctx context.Context, args ...string) (string, error) {
	address := os.Getenv("REDIS_ADDR")
	if address == "" {
		address = "localhost:6379"
	}
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(4 * time.Second))
	reader := bufio.NewReader(conn)
	if p := os.Getenv("REDIS_PASSWORD"); p != "" {
		if err := send(conn, "AUTH", p); err != nil {
			return "", err
		}
		if _, err := readRESP(reader); err != nil {
			return "", err
		}
	}
	if err := send(conn, args...); err != nil {
		return "", err
	}
	return readRESP(reader)
}
func send(w io.Writer, args ...string) error {
	_, err := fmt.Fprintf(w, "*%d\r\n", len(args))
	if err != nil {
		return err
	}
	for _, a := range args {
		if _, err = fmt.Fprintf(w, "$%d\r\n%s\r\n", len(a), a); err != nil {
			return err
		}
	}
	return nil
}
func readRESP(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if len(line) == 0 {
		return "", errors.New("empty redis response")
	}
	switch line[0] {
	case '+', ':':
		return line[1:], nil
	case '-':
		return "", errors.New(line[1:])
	case '$':
		size, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", err
		}
		if size < 0 {
			return "", nil
		}
		buf := make([]byte, size+2)
		_, err = io.ReadFull(r, buf)
		return string(buf[:size]), err
	default:
		return "", errors.New("unsupported redis response")
	}
}
func Revoked(ctx context.Context, jti string) (bool, error) {
	v, err := redisCommand(ctx, "EXISTS", "jwt:blacklist:"+jti)
	return v == "1", err
}
func Revoke(ctx context.Context, c Claims) error {
	ttl := c.ExpiresAt - time.Now().Unix()
	if ttl <= 0 {
		return nil
	}
	_, err := redisCommand(ctx, "SET", "jwt:blacklist:"+c.ID, "1", "EX", strconv.FormatInt(ttl, 10))
	return err
}
func Ping(ctx context.Context) error { _, err := redisCommand(ctx, "PING"); return err }
