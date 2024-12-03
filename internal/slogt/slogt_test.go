package slogt_test

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/neilotoole/slogt"
)

const (
	iter  = 3
	sleep = time.Millisecond * 20
)

func TestSlog_Ugly(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	t.Log("I am indented correctly")
	log.Info("But I am not")
}

func TestSlogt_Pretty(t *testing.T) {
	log := slogt.New(t)
	t.Log("I am indented correctly")
	log.Info("And so am I")
}

// TestSlog_Ugly_Parallel demonstrates that testing output is particularly
// ugly when using t.Parallel(), because
// the slog.Logger output is not tied to the testing.T.
func TestSlog_Ugly_Parallel(t *testing.T) {
	for i := 0; i < iter; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			handler := slog.NewTextHandler(os.Stdout, nil)
			log := slog.New(handler)

			for j := 0; j < iter; j++ {
				t.Log("YAY: this is indented as expected.")
				log := log.With("count", j)
				log.Info("BOO: This, alas, is not indented.")

				// Sleep a little to allow the goroutines to interleave.
				time.Sleep(sleep)
			}
		})
	}
}

// TestSlogt_Pretty demonstrates use of slog with testing.T.
func TestSlogt_Pretty_Parallel(t *testing.T) {
	for i := 0; i < iter; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			log := slogt.New(t)
			for j := 0; j < iter; j++ {
				t.Log("testing.T: this is indented as expected.")

				log.Debug("slogt: debug")
				log.Info("slogt: info")
				log = log.With("count", j)
				log.Info("slogt: info with attrs")

				// Sleep a little to allow the goroutines to interleave.
				time.Sleep(sleep)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	log := slogt.New(t)
	log.Debug("debug me")
	log.Info("info me")
	log.Warn("warn me")
	log.Error("error me")
}

func TestText(t *testing.T) {
	log := slogt.New(t, slogt.Text())
	log.Info("hello world")
}

func TestJSON(t *testing.T) {
	log := slogt.New(t, slogt.JSON())
	log.Info("hello world")
}

func TestFactory(t *testing.T) {
	// This factory returns a slog.Handler using slog.LevelError.
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		opts := &slog.HandlerOptions{
			Level: slog.LevelError,
		}
		return slog.NewTextHandler(w, opts)
	})

	log := slogt.New(t, f)
	log.Info("Should NOT be printed because level is too low")
	log.Error("Should be printed because level is sufficiently high")
}

func TestCaller(t *testing.T) {
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		opts := &slog.HandlerOptions{
			AddSource: true,
		}

		return slog.NewTextHandler(w, opts)
	})

	log := slogt.New(t, f)
	log.Info("Show me the real callsite")
}

func TestDefaultHandler(t *testing.T) {
	slogt.Default = slogt.JSON()
	log := slogt.New(t)
	log.Info("should show json")
}
