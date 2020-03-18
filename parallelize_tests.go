// Copyright 2020 Google LLC.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/orijtech/otils"
)

func main() {
	testApps := otils.NonEmptyStrings(strings.Split(os.Getenv("APPS"), " ")...)
	if len(testApps) == 0 {
		panic("No APPS passed in")
	}

	rand.Shuffle(len(testApps), func(i, j int) {
		testApps[i], testApps[j] = testApps[j], testApps[i]
	})

	// Otherwise, we'll parallelize the builds.
	nParallel := runtime.GOMAXPROCS(0)
	println(nParallel)

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	stride := 1 + (len(testApps) / nParallel)
	for i := 0; i < len(testApps); i += stride {
		apps := testApps[i : i+stride]
		wg.Add(1)
		go runTests(shutdownCtx, &wg, apps, "script.sh")
	}
}

func runTests(ctx context.Context, wg *sync.WaitGroup, apps []string, testSuiteScriptPath string) error {
	defer wg.Done()

	if len(apps) == 0 {
		return errors.New("Expected at least one app")
	}

	cmd := exec.CommandContext(ctx, "bash", testSuiteScriptPath)
	cmd.Env = append(os.Environ(), `TEST_APPS=`+strings.Join(apps, " ")+``)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	return nil
}
