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

package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	// Scenarios can contain one of many of the following reserved
	// keywords that are automatically populated.

	// Sender is the sender and signer of a transaction.
	Sender = "{{ SENDER }}"

	// SenderValue is the amount the sender is paying.
	SenderValue = "{{ SENDER_VALUE }}"

	// Recipient is the recipient of the transaction.
	Recipient = "{{ RECIPIENT }}"

	// RecipientValue is the amount the recipient is
	// receiving from the sender. Note, this is distinct
	// from the SenderValue so that UTXO transfers
	// can be supported.
	RecipientValue = "{{ RECIPIENT_VALUE }}"

	// UTXOIdentifier is the globally unique identifier
	// of a UTXO. This should be in the Operation.metadata
	// of any UTXO-based blockchain ("utxo_created" when
	// a new UTXO is created and "utxo_spent" when a
	// UTXO is spent).
	UTXOIdentifier = "{{ UTXO_IDENTIFIER }}"
)

// Context is all information passed to PopulateScenario.
// As more exotic scenario testing is supported, this will
// likely be expanded.
type Context struct {
	Sender         string
	SenderValue    *big.Int
	Recipient      string
	RecipientValue *big.Int
	UTXOIdentifier string
	Currency       *types.Currency
}

// PopulateScenario populates a provided scenario (slice of
// []*types.Operation) with the information in Context.
func PopulateScenario(
	ctx context.Context,
	scenarioContext *Context,
	scenario []*types.Operation,
) ([]*types.Operation, error) {
	// Convert operations to a string
	bytes, err := json.Marshal(scenario)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to marshal scenario", err)
	}

	// Replace all keywords with information in Context
	stringBytes := string(bytes)
	stringBytes = strings.ReplaceAll(stringBytes, Sender, scenarioContext.Sender)
	stringBytes = strings.ReplaceAll(
		stringBytes,
		SenderValue,
		new(big.Int).Neg(scenarioContext.SenderValue).String(),
	)
	stringBytes = strings.ReplaceAll(stringBytes, Recipient, scenarioContext.Recipient)
	stringBytes = strings.ReplaceAll(
		stringBytes,
		RecipientValue,
		new(big.Int).Abs(scenarioContext.RecipientValue).String(),
	)

	if len(scenarioContext.UTXOIdentifier) > 0 {
		stringBytes = strings.ReplaceAll(
			stringBytes,
			UTXOIdentifier,
			scenarioContext.UTXOIdentifier,
		)
	}

	// Convert back to ops
	var ops []*types.Operation
	if err := json.Unmarshal([]byte(stringBytes), &ops); err != nil {
		return nil, fmt.Errorf("%w: unable to unmarshal ops", err)
	}

	// Post-process operations
	for _, op := range ops {
		if op.Amount != nil {
			op.Amount.Currency = scenarioContext.Currency
		}
	}

	return ops, nil
}
