// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/fatih/color"
)

const (
	// DefaultFilePermissions specifies that the user can
	// read and write the file.
	DefaultFilePermissions = 0600

	// AllFilePermissions specifies anyone can do anything
	// to the file.
	AllFilePermissions = 0777
)

// CreateTempDir creates a directory in
// /tmp for usage within testing.
func CreateTempDir() (string, error) {
	storageDir, err := ioutil.TempDir("", "rosetta-cli")
	if err != nil {
		return "", err
	}

	color.Cyan("Using temporary directory %s", storageDir)
	return storageDir, nil
}

// RemoveTempDir deletes a directory at
// a provided path for usage within testing.
func RemoveTempDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		log.Fatal(err)
	}
}

// EnsurePathExists creates directories along
// a path if they do not exist.
func EnsurePathExists(path string) error {
	if err := os.MkdirAll(path, os.FileMode(AllFilePermissions)); err != nil {
		return fmt.Errorf("%w: unable to create data and network directory", err)
	}

	return nil
}

// Equal returns a boolean indicating if two
// interfaces are equal.
func Equal(a interface{}, b interface{}) bool {
	return types.Hash(a) == types.Hash(b)
}

// SerializeAndWrite attempts to serialize the provided object
// into a file at filePath.
func SerializeAndWrite(filePath string, object interface{}) error {
	err := ioutil.WriteFile(
		filePath,
		[]byte(types.PrettyPrintStruct(object)),
		os.FileMode(DefaultFilePermissions),
	)
	if err != nil {
		return fmt.Errorf("%w: unable to write to file path %s", err, filePath)
	}

	return nil
}

// LoadAndParse reads the file at the provided path
// and attempts to unmarshal it into output.
func LoadAndParse(filePath string, output interface{}) error {
	bytes, err := ioutil.ReadFile(path.Clean(filePath))
	if err != nil {
		return fmt.Errorf("%w: unable to load file %s", err, filePath)
	}

	if err := json.Unmarshal(bytes, &output); err != nil {
		return fmt.Errorf("%w: unable to unmarshal", err)
	}

	return nil
}

// CreateCommandPath creates a unique path for a command and network within a data directory. This
// is used to avoid collision when using multiple commands on multiple networks
// when the same storage resources are used. If the derived path does not exist,
// we run os.MkdirAll on the path.
func CreateCommandPath(
	dataDirectory string,
	cmd string,
	network *types.NetworkIdentifier,
) (string, error) {
	dataPath := path.Join(dataDirectory, cmd, types.Hash(network))
	if err := EnsurePathExists(dataPath); err != nil {
		return "", fmt.Errorf("%w: cannot populate path", err)
	}

	return dataPath, nil
}

// CheckNetworkSupported checks if a Rosetta implementation supports a given
// *types.NetworkIdentifier. If it does, the current network status is returned.
func CheckNetworkSupported(
	ctx context.Context,
	networkIdentifier *types.NetworkIdentifier,
	fetcher *fetcher.Fetcher,
) (*types.NetworkStatusResponse, error) {
	networks, err := fetcher.NetworkList(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to fetch network list", err)
	}

	networkMatched := false
	for _, availableNetwork := range networks.NetworkIdentifiers {
		if types.Hash(availableNetwork) == types.Hash(networkIdentifier) {
			networkMatched = true
			break
		}
	}

	if !networkMatched {
		return nil, fmt.Errorf("%s is not available", types.PrettyPrintStruct(networkIdentifier))
	}

	status, err := fetcher.NetworkStatusRetry(
		ctx,
		networkIdentifier,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to get network status", err)
	}

	return status, nil
}
