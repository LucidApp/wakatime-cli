package heartbeat_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestSendHeartbeats(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &[]interface{}{&entity})
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(string(expectedBody), entity.Entity, heartbeat.UserAgent(plugin))

		assert.True(t, strings.HasSuffix(entity.Entity, "testdata/main.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_WithFiltering_Exclude(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("entity", `\tmp\main.go`)
	v.Set("exclude", `/tmp/`)
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin")
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 0, numCalls)
}

func TestSendHeartbeats_ExtraHeartbeats(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_extra_heartbeats_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[2].Entity, "testdata/main.py"))

		for i := 3; i < 25; i++ {
			assert.True(t, strings.HasSuffix(entities[i].Entity, "testdata/main.go"))
		}

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, heartbeat.UserAgent(plugin),
			entities[1].Entity, heartbeat.UserAgent(plugin),
			entities[2].Entity, heartbeat.UserAgent(plugin),
			entities[3].Entity, heartbeat.UserAgent(plugin),
			entities[4].Entity, heartbeat.UserAgent(plugin),
			entities[5].Entity, heartbeat.UserAgent(plugin),
			entities[6].Entity, heartbeat.UserAgent(plugin),
			entities[7].Entity, heartbeat.UserAgent(plugin),
			entities[8].Entity, heartbeat.UserAgent(plugin),
			entities[9].Entity, heartbeat.UserAgent(plugin),
			entities[10].Entity, heartbeat.UserAgent(plugin),
			entities[11].Entity, heartbeat.UserAgent(plugin),
			entities[12].Entity, heartbeat.UserAgent(plugin),
			entities[13].Entity, heartbeat.UserAgent(plugin),
			entities[14].Entity, heartbeat.UserAgent(plugin),
			entities[15].Entity, heartbeat.UserAgent(plugin),
			entities[16].Entity, heartbeat.UserAgent(plugin),
			entities[17].Entity, heartbeat.UserAgent(plugin),
			entities[18].Entity, heartbeat.UserAgent(plugin),
			entities[19].Entity, heartbeat.UserAgent(plugin),
			entities[20].Entity, heartbeat.UserAgent(plugin),
			entities[21].Entity, heartbeat.UserAgent(plugin),
			entities[22].Entity, heartbeat.UserAgent(plugin),
			entities[23].Entity, heartbeat.UserAgent(plugin),
			entities[24].Entity, heartbeat.UserAgent(plugin),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 0)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 1)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("lineno", 2)
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ExtraHeartbeats_Sanitize(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 0)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hide-branch-names", true)
	v.Set("hide-file-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("lineno", 13)
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	db, err := bolt.Open(offlineQueueFile.Name(), 0600, nil)
	require.NoError(t, err)

	defer db.Close()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)

	hh, err := q.PopMany(1)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)
	assert.Len(t, hh, 1)

	info, err := goInfo.GetInfo()
	require.NoError(t, err)

	userAgent := fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s %s",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
		plugin,
	)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Branch:         nil,
			Category:       heartbeat.CodingCategory,
			CursorPosition: nil,
			Dependencies:   nil,
			Entity:         "HIDDEN.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.PointerTo(true),
			Language:       heartbeat.PointerTo("Go"),
			LineNumber:     nil,
			Lines:          nil,
			Project:        heartbeat.PointerTo("wakatime-cli"),
			Time:           1585598059,
			UserAgent:      userAgent,
		}}, hh)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_NonExistingEntity(t *testing.T) {
	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", "https://example.org")
	v.Set("entity", "nonexisting")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())

	cmd.SetupLogging(v)

	defer func() {
		if file, ok := log.Output().(*os.File); ok {
			_ = file.Sync()
			file.Close()
		} else if handler, ok := log.Output().(io.Closer); ok {
			handler.Close()
		}
	}()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	err = cmdheartbeat.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestSendHeartbeats_IsUnsavedEntity(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_is_unsaved_entity_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "missing"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "missing-from-extra-heartbeats"))
		assert.True(t, strings.HasSuffix(entities[2].Entity, "main.go"))

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, heartbeat.UserAgent(plugin),
			entities[1].Entity, heartbeat.UserAgent(plugin),
			entities[2].Entity, heartbeat.UserAgent(plugin),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_is_unsaved_entity.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	inr, inw, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		inr.Close()
		inw.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = inr

	data, err := os.ReadFile("testdata/extra_heartbeats_is_unsaved_entity.json")
	require.NoError(t, err)

	go func() {
		_, err := inw.Write(data)
		require.NoError(t, err)

		inw.Close()
	}()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("is-unsaved-entity", true)
	v.Set("category", "coding")
	v.Set("cursorpos", 41)
	v.Set("entity", "missing")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("project", "wakatime-cli")
	v.Set("hide-branch-names", true)
	v.Set("lineno", 11)
	v.Set("lines-in-file", 91)
	v.Set("plugin", plugin)
	v.Set("time", 1585598051)
	v.Set("timeout", 5)
	v.Set("extra-heartbeats", true)
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	cmd.SetupLogging(v)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer func() {
		offlineQueueFile.Close()
		logFile.Close()

		if file, ok := log.Output().(*os.File); ok {
			_ = file.Sync()
			file.Close()
		} else if handler, ok := log.Output().(io.Closer); ok {
			handler.Close()
		}
	}()

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestSendHeartbeats_NonExistingExtraHeartbeatsEntity(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_extra_heartbeats_filtered_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "testdata/main.py"))

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, heartbeat.UserAgent(plugin),
			entities[1].Entity, heartbeat.UserAgent(plugin),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats_filtered.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	inr, inw, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		inr.Close()
		inw.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = inr

	data, err := os.ReadFile("testdata/extra_heartbeats_nonexisting_entity.json")
	require.NoError(t, err)

	go func() {
		_, err := inw.Write(data)
		require.NoError(t, err)

		inw.Close()
	}()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	cmd.SetupLogging(v)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer func() {
		offlineQueueFile.Close()
		logFile.Close()

		if file, ok := log.Output().(*os.File); ok {
			_ = file.Sync()
			file.Close()
		} else if handler, ok := log.Output().(io.Closer); ok {
			handler.Close()
		}
	}()

	err = cmdheartbeat.SendHeartbeats(v, offlineQueueFile.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
